# tg-bot-movieclub-go

A Telegram bot for managing a movie club.

## Project Structure

```
.
├── cmd/
│   └── bot/           # Main application entry point
├── pkg/
│   └── bot/           # Public bot package
├── internal/
│   └── config/        # Internal configuration
└── go.mod             # Go module file
```

## Prerequisites

- Go 1.24 or later

## Building

```bash
go build -o bin/bot ./cmd/bot
```

## Running

```bash
./bin/bot
```

Or directly with Go:

```bash
go run ./cmd/bot
```

## Configuration

Configuration is managed through environment variables:

- `TELEGRAM_TOKEN` - Your Telegram bot token

## Development

### Running Tests

```bash
go test ./...
```

### Formatting Code

```bash
go fmt ./...
```

### Linting

```bash
go vet ./...
```