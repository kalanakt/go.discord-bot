#!/bin/bash

# Development helper script for Go Discord Bot

set -e

function check_env_file() {
    if [ ! -f .env ]; then
        echo "Error: .env file not found."
        echo "Creating .env file from .env.example..."
        cp .env.example .env
        echo "Please edit the .env file with your configuration."
        exit 1
    fi
}

function build() {
    echo "Building Discord Bot..."
    go build -o discord-bot .
    echo "Build complete."
}

function run() {
    check_env_file
    echo "Running Discord Bot..."
    go run main.go
}

function docker_up() {
    check_env_file
    echo "Starting Docker containers..."
    docker-compose up -d
    echo "Docker containers started. Bot is running."
}

function docker_down() {
    echo "Stopping Docker containers..."
    docker-compose down
    echo "Docker containers stopped."
}

function docker_logs() {
    echo "Showing Docker logs..."
    docker-compose logs -f
}

function migrate() {
    check_env_file
    echo "Running database migrations..."
    source .env
    go run main.go -migrate-only
    echo "Migrations complete."
}

function help() {
    echo "Go Discord Bot Development Helper"
    echo ""
    echo "Usage:"
    echo "  ./dev.sh [command]"
    echo ""
    echo "Commands:"
    echo "  build       - Build the Discord bot"
    echo "  run         - Run the Discord bot locally"
    echo "  docker-up   - Start Docker containers"
    echo "  docker-down - Stop Docker containers"
    echo "  docker-logs - Show Docker container logs"
    echo "  migrate     - Run database migrations"
    echo "  help        - Show this help message"
}

# Main script execution
if [ $# -eq 0 ]; then
    help
    exit 0
fi

case "$1" in
    build)
        build
        ;;
    run)
        run
        ;;
    docker-up)
        docker_up
        ;;
    docker-down)
        docker_down
        ;;
    docker-logs)
        docker_logs
        ;;
    migrate)
        migrate
        ;;
    help)
        help
        ;;
    *)
        echo "Unknown command: $1"
        help
        exit 1
        ;;
esac