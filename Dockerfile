# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set Go environment variables for better reliability
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies with retry logic
RUN go mod download || \
    (sleep 5 && go mod download) || \
    (sleep 10 && go mod download)

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o discord-bot .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata ffmpeg

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/discord-bot /app/discord-bot

# Copy database migrations
COPY --from=builder /app/database/migrations /app/database/migrations

# Run the application
CMD ["/app/discord-bot"]