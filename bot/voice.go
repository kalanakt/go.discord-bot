package bot

import (
	"errors"
	"io"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// VoiceConnection represents a voice connection to a Discord guild
type VoiceConnection struct {
	GuildID   string
	ChannelID string
	Conn      *discordgo.VoiceConnection
	Playing   bool
	Stopping  bool
	Mu        sync.Mutex
}

// VoiceManager manages voice connections across guilds
type VoiceManager struct {
	Bot         *Bot
	Connections map[string]*VoiceConnection // Map of guild ID to voice connection
	Mu          sync.Mutex
}

// NewVoiceManager creates a new voice manager
func NewVoiceManager(bot *Bot) *VoiceManager {
	return &VoiceManager{
		Bot:         bot,
		Connections: make(map[string]*VoiceConnection),
	}
}

// JoinVoiceChannel joins a voice channel
func (vm *VoiceManager) JoinVoiceChannel(guildID, channelID string) (*VoiceConnection, error) {
	vm.Mu.Lock()
	defer vm.Mu.Unlock()

	// Check if already connected to this guild
	if vc, ok := vm.Connections[guildID]; ok {
		// If already in the requested channel, return the existing connection
		if vc.ChannelID == channelID {
			return vc, nil
		}

		// Otherwise, disconnect from the current channel
		if err := vc.Conn.Disconnect(); err != nil {
			logrus.Warnf("Error disconnecting from voice channel: %v", err)
		}
		delete(vm.Connections, guildID)
	}

	// Join the new voice channel
	conn, err := vm.Bot.Session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	// Create and store the voice connection
	vc := &VoiceConnection{
		GuildID:   guildID,
		ChannelID: channelID,
		Conn:      conn,
		Playing:   false,
		Stopping:  false,
	}
	vm.Connections[guildID] = vc

	return vc, nil
}

// LeaveVoiceChannel leaves a voice channel
func (vm *VoiceManager) LeaveVoiceChannel(guildID string) error {
	vm.Mu.Lock()
	defer vm.Mu.Unlock()

	// Check if connected to this guild
	vc, ok := vm.Connections[guildID]
	if !ok {
		return errors.New("not connected to a voice channel in this guild")
	}

	// Stop any playing audio
	vc.Mu.Lock()
	vc.Stopping = true
	vc.Mu.Unlock()

	// Disconnect from the voice channel
	if err := vc.Conn.Disconnect(); err != nil {
		return err
	}

	// Remove the connection from the map
	delete(vm.Connections, guildID)

	return nil
}

// PlayAudio plays audio from a reader
func (vc *VoiceConnection) PlayAudio(reader io.Reader) error {
	vc.Mu.Lock()
	if vc.Playing {
		vc.Mu.Unlock()
		return errors.New("already playing audio")
	}
	vc.Playing = true
	vc.Stopping = false
	vc.Mu.Unlock()

	// Make sure we're speaking
	if err := vc.Conn.Speaking(true); err != nil {
		return err
	}

	// When we're done, stop speaking and set playing to false
	defer func() {
		_ = vc.Conn.Speaking(false)
		vc.Mu.Lock()
		vc.Playing = false
		vc.Stopping = false
		vc.Mu.Unlock()
	}()

	// Create a buffer for audio data
	buf := make([]byte, 16*1024) // 16KB buffer
	for {
		// Check if we should stop
		vc.Mu.Lock()
		if vc.Stopping {
			vc.Mu.Unlock()
			break
		}
		vc.Mu.Unlock()

		// Read from the audio source
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				logrus.Errorf("Error reading audio: %v", err)
			}
			break
		}

		// Send the audio data to Discord
		if n > 0 {
			vc.Conn.OpusSend <- buf[:n]
		}
	}

	return nil
}

// StopAudio stops playing audio
func (vc *VoiceConnection) StopAudio() {
	vc.Mu.Lock()
	defer vc.Mu.Unlock()

	vc.Stopping = true
}

// IsPlaying returns whether audio is currently playing
func (vc *VoiceConnection) IsPlaying() bool {
	vc.Mu.Lock()
	defer vc.Mu.Unlock()

	return vc.Playing
}