# 🎬 Movie Club Telegram Bot

A comprehensive Telegram bot for managing a movie club with voting system, movie suggestions, and session management.

## ✨ Features

- 🎥 **Movie Management**: Suggest, track, and manage movies
- 🗳️ **Voting System**: 
  - Selection voting (choose next movie to watch)
  - Rating voting (rate movies after watching with 1-10 scale)
- 📅 **Session Management**: Track movie viewing sessions
- 👥 **User Roles**: Admin and member permissions
- 📊 **Telegraph Integration**: Generate beautiful movie lists
- 🔄 **State Machine**: FSM-based conversation flows
- 🎯 **Kinopoisk API**: Fetch movie information from Kinopoisk
- ⏰ **Background Tasks**: Scheduled voting closures with Asynq + Redis

## 🏗️ Project Structure

```
.
├── cmd/
│   └── bot/                    # Main application entry point
├── internal/
│   ├── app/                    # Application initialization
│   ├── config/                 # Configuration management
│   ├── db/                     # Database setup and migrations
│   ├── model/                  # Data models (GORM)
│   │   ├── movie.go
│   │   ├── session.go
│   │   ├── user.go
│   │   ├── voting.go
│   │   ├── vote.go
│   │   ├── poll.go
│   │   └── poll_option.go
│   ├── repository/             # Database repositories
│   │   ├── movie_repo.go
│   │   ├── session_repo.go
│   │   ├── vote_repo.go
│   │   └── voting_repo.go
│   ├── service/                # Business logic
│   │   ├── movie_service.go
│   │   ├── voting_service.go
│   │   ├── vote_service.go
│   │   └── kinopoisk_service.go
│   ├── transport/telegram/     # Telegram handlers
│   │   ├── voting.go
│   │   ├── suggest_movie.go
│   │   ├── cancel_voting.go
│   │   ├── current_movies.go
│   │   └── already_watched_movies.go
│   ├── tasks/                  # Background tasks
│   │   ├── close_selection_voting.go
│   │   └── close_rating_voting.go
│   └── utils/                  # Utilities
│       ├── fsm/                # Finite State Machine
│       ├── kinopoisk/          # Kinopoisk API client
│       ├── telegram/           # Telegram utilities
│       │   ├── keyboard/       # Inline keyboards
│       │   └── middleware/     # Auth & permissions
│       └── telegraph/          # Telegraph integration
├── scripts/
│   ├── export_movies/          # Export movies to JSON
│   └── import_movies/          # Import movies from JSON
└── go.mod
```

## 📋 Prerequisites

- Go 1.25.3 or later
- SQLite3
- Redis (for background tasks)
- Telegram Bot Token
- Kinopoisk API Key (optional)

## 🚀 Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Forceres/tg-bot-movieclub-go.git
   cd tg-bot-movieclub-go
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment**
   
   Create a `.env` file:
   ```env
   # Telegram
   TELEGRAM_BOT_TOKEN=your_bot_token_here
	 TELEGRAM_WEBHOOK_SECRET_TOKEN=your_webhook_secret
   TELEGRAM_GROUP_ID=your_group_id_here
   
   # Database
   DATABASE_NAME=db.sqlite3
   
   # Redis
   REDIS_URL=your_redis_url
   
   # Kinopoisk API (optional)
   KINOPOISK_API_KEY=your_api_key_here
   KINOPOISK_API_URL=https://kinopoiskapiunofficial.tech/api
	 KINOPOISK_API_VERSION=
   ```

4. **Run migrations**
   
   Migrations run automatically on first start, creating:
   - `users` table
   - `movies` table
   - `sessions` table
   - `votings` table
   - `votes` table
   - `polls` table
   - `poll_options` table
   - `roles` table (with admin/user roles seeded)

## 🎯 Usage

### Running the bot

```bash
# Using Make
make run

# Or directly with Go
go run ./cmd/bot

# Build and run
make build
./bin/bot
```

### Bot Commands

- `/start` - Start the bot
- `/help` - Show help message
- `/voting` - Create a new voting (admin only)
- `/suggest` - Suggest a movie to watch
- `/current` - Show current movies
- `/watched` - Show already watched movies
- `/cancel` - Cancel current operation
- `/cancel_voting` - Cancel active voting (admin only)

### Creating Votings

1. **Selection Voting** (Choose next movie):
   - Admin creates voting with title and duration
   - Members vote for their preferred movie
   - Movie with most votes wins

2. **Rating Voting** (Rate watched movie):
   - Admin creates rating voting for specific movie(s)
   - Members rate on scale 1-10
   - Average rating is calculated

## 🛠️ Development

### Running Tests

```bash
make test
```

### Formatting Code

```bash
make fmt
```

### Linting

```bash
make vet
```

### Run all checks

```bash
make check
```

## 📊 Database Schema

### Core Tables

- **users**: Telegram users with roles
- **movies**: Movie information (title, year, IMDB, poster, etc.)
- **sessions**: Movie viewing sessions (many-to-many with movies)
- **votings**: Voting sessions (selection or rating type)
- **votes**: Individual votes within votings
- **polls**: Telegram poll tracking (survives bot restarts)
- **poll_options**: Maps poll options to movies

### Relationships

```
users ←→ votings (creator)
users ←→ votes (voter)
movies ←→ sessions (many-to-many)
movies ←→ votes (voted movie)
votings ←→ votes (one-to-many)
votings ←→ polls (one-to-many)
polls ←→ poll_options (one-to-many)
poll_options → movies
```

## 🔧 Utilities

### Export Movies

```bash
go run scripts/export_movies/export_movies.go
```

### Import Movies

```bash
go run scripts/import_movies/import_movies.go -file movies.json
```

## 🏗️ Architecture

### Layered Architecture

1. **Transport Layer** (`transport/telegram/`): Handles Telegram updates
2. **Service Layer** (`service/`): Business logic
3. **Repository Layer** (`repository/`): Data access
4. **Model Layer** (`model/`): Data structures

### State Management

Uses FSM (Finite State Machine) for conversation flows:
- Default state
- Voting creation flow
- Movie suggestion flow
- etc.

### Background Tasks

Uses Asynq + Redis for scheduled tasks:
- Auto-close votings when time expires
- Send reminders
- Cleanup old data

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 👤 Author

**Forceres**

- GitHub: [@Forceres](https://github.com/Forceres)

## 🙏 Acknowledgments

- [go-telegram/bot](https://github.com/go-telegram/bot) - Telegram Bot API wrapper
- [GORM](https://gorm.io/) - ORM library
- [Asynq](https://github.com/hibiken/asynq) - Task queue
- [Telegraph](https://telegra.ph/) - Content publishing platform