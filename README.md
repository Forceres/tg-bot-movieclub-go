# 🎬 Movie Club Telegram Bot

A comprehensive Telegram bot for managing a movie club with voting system, movie suggestions, session management, and scheduling. Built with Go, featuring automatic task scheduling, multi-stage Docker deployment, and rich user interactions with emoji-enhanced messages.

## ✨ Features

### 🎥 Movie Management
- **Movie Suggestions**: Suggest movies via Kinopoisk links or IDs
- **Session Management**: Create and manage viewing sessions with multiple movies
- **Movie Tracking**: Track suggested vs watched movies
- **Custom Descriptions**: Add custom descriptions to viewing sessions
- **Automatic Info Fetching**: Get movie details from Kinopoisk API automatically

### 🗳️ Voting System
- **Selection Voting**: Choose next movie to watch from suggestions
- **Rating Voting**: Rate movies after watching (1-10 scale)
- **Automated Scheduling**: Votings auto-close after specified duration
- **Poll Persistence**: Polls tracked in database (survives bot restarts)
- **Vote Tracking**: Complete vote history per user

### 📅 Session & Schedule Management
- **Session Creation**: Automatically create sessions when adding movies
- **Session Rescheduling**: Change session dates/times
- **Session Cancellation**: Cancel sessions with automatic cleanup
- **Recurring Schedules**: Set weekly schedules for movie nights
- **Automatic Task Scheduling**: Auto-schedule rating votings and session completions

### 👥 User Management
- **Auto-Registration**: Users automatically registered on first interaction
- **Role-Based Access**: Admin and member roles with different permissions
- **Permission Middleware**: Group membership and admin verification
- **User Tracking**: Track who suggested movies, created sessions

### 🎨 User Experience
- **Emoji-Enhanced Messages**: All user-facing messages use contextual emoji
- **Interactive Keyboards**: Inline keyboards for selections
- **Date Picker**: Visual date selection for scheduling
- **Paginated Lists**: Paginated movie and voting lists
- **State Management**: FSM-based conversation flows for complex interactions

### 🔧 Technical Features
- **Background Tasks**: Asynq + Redis for scheduled task execution
- **Telegraph Integration**: Generate beautiful shareable movie lists
- **Docker Support**: Multi-stage builds for bot and worker
- **Database Migrations**: Automatic GORM migrations
- **Error Handling**: Comprehensive error handling with user-friendly messages

## 🏗️ Project Structure

```
tg-bot-movieclub-go/
├── cmd/
│   ├── bot/                    # Bot application entry point
│   │   └── main.go
│   └── worker/                 # Worker application entry point
│       └── main.go
├── internal/
│   ├── app/                    # Application initialization
│   │   └── app.go
│   ├── config/                 # Configuration management
│   │   ├── app.go
│   │   ├── config.go
│   │   ├── db.go
│   │   ├── kinopoisk.go
│   │   ├── redis.go
│   │   └── telegram.go
│   ├── db/                     # Database setup and migrations
│   │   └── db.go
│   ├── model/                  # Data models (GORM)
│   │   ├── movie.go            # Movie entity
│   │   ├── session.go          # Viewing session
│   │   ├── user.go             # User and role
│   │   ├── voting.go           # Voting entity
│   │   ├── vote.go             # Individual vote
│   │   ├── poll.go             # Telegram poll tracking
│   │   ├── poll_option.go      # Poll option mapping
│   │   └── schedule.go         # Recurring schedule
│   ├── repository/             # Database repositories
│   │   ├── movie_repo.go
│   │   ├── session_repo.go
│   │   ├── user_repo.go
│   │   ├── vote_repo.go
│   │   ├── voting_repo.go
│   │   ├── poll_repo.go
│   │   ├── role_repo.go
│   │   └── schedule_repo.go
│   ├── service/                # Business logic layer
│   │   ├── movie_service.go
│   │   ├── session_service.go
│   │   ├── user_service.go
│   │   ├── voting_service.go
│   │   ├── vote_service.go
│   │   ├── kinopoisk_service.go
│   │   ├── poll_service.go
│   │   └── schedule_service.go
│   ├── transport/telegram/     # Telegram handlers
│   │   ├── add_movie_to_session.go      # /adds command
│   │   ├── remove_movie_from_session.go # /removes command
│   │   ├── cancel_session.go            # /cancel_session command
│   │   ├── reshedule_session.go         # /reschedule command
│   │   ├── custom_session_description.go # /custom command
│   │   ├── voting.go                    # /voting command
│   │   ├── cancel_voting.go             # /cancel_voting command
│   │   ├── suggest_movie.go             # Movie suggestion handler
│   │   ├── current_movies.go            # /current command
│   │   ├── already_watched_movies.go    # /watched command
│   │   ├── schedule.go                  # /schedule command
│   │   ├── poll_answer.go               # Poll answer handler
│   │   ├── register_user.go             # User registration
│   │   ├── update_chat_member.go        # Member updates
│   │   ├── cancel.go                    # /cancel command
│   │   ├── help.go                      # /help command
│   │   └── default.go                   # Default message handler
│   ├── tasks/                  # Background tasks (Asynq)
│   │   ├── queue.go                     # Queue setup
│   │   ├── finish_session.go            # Session completion task
│   │   ├── open_rating_voting.go        # Rating voting task
│   │   ├── close_selection_voting.go    # Selection voting closure
│   │   └── close_rating_voting.go       # Rating voting closure
│   └── utils/                  # Utilities
│       ├── date/               # Date utilities
│       │   └── date.go
│       ├── fsm/                # Finite State Machine
│       │   └── fsm.go
│       ├── kinopoisk/          # Kinopoisk API client
│       │   ├── api.go
│       │   └── parse.go
│       ├── slice/              # Slice utilities
│       │   └── slice.go
│       ├── telegram/           # Telegram utilities
│       │   ├── datepicker/     # Date picker widget
│       │   ├── keyboard/       # Inline keyboards
│       │   └── middleware/     # Auth & permissions
│       └── telegraph/          # Telegraph integration
│           └── init.go
├── scripts/
│   ├── export_movies/          # Export movies to JSON
│   │   └── export_movies.go
│   ├── import_movies/          # Import movies from JSON
│   │   └── import_movies.go
│   └── json_to_sql/            # Generate SQL from JSON
│       └── json_to_sql.go
├── build/                      # Build artifacts directory
├── Dockerfile                  # Multi-stage Docker build
├── docker-compose.yml          # Development compose
├── docker-compose.prod.yml     # Production compose with nginx
├── Makefile                    # Build automation
├── movies.json                 # Movie database export
├── movies_insert.sql           # SQL insert statements
└── go.mod                      # Go dependencies
```

## �️ Technology Stack

### Core
- **Go 1.25.3** - Programming language
- **Postgresql** - Database (via GORM)
- **Redis** - Task queue backend
- **Docker** - Containerization

### Libraries & Frameworks
- **[go-telegram/bot](https://github.com/go-telegram/bot)** v1.17.0 - Telegram Bot API
- **[GORM](https://gorm.io/)** v1.31.1 - ORM for database operations
- **[Asynq](https://github.com/hibiken/asynq)** v0.25.1 - Task queue and scheduler
- **[FSM](https://github.com/go-telegram/fsm)** v0.2.0 - Finite State Machine
- **[Telegraph Go](https://github.com/celestix/telegraph-go)** v2.0.4 - Telegraph API client
- **[cleanenv](https://github.com/ilyakaznacheev/cleanenv)** v1.5.0 - Configuration management

### Development Tools
- **Make** - Build automation
- **Docker Compose** - Local development orchestration

## �📋 Prerequisites

- **Go** 1.25.3 or later
- **Postgresql**
- **Redis** 6.0+ (for background tasks)
- **Docker** & **Docker Compose** (optional, for containerized deployment)
- **Telegram Bot Token** (from @BotFather)
- **Kinopoisk API Key** (from kinopoiskapiunofficial.tech)

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
   
   Create a `.env` file in the project root:
   ```env
   # Telegram Bot Configuration
   TELEGRAM_BOT_TOKEN=your_bot_token_here
   TELEGRAM_WEBHOOK_SECRET_TOKEN=your_webhook_secret_token
   TELEGRAM_GROUP_ID=-100123456789  # Your group chat ID (negative for supergroups)
   
   # Database Configuration
   DATABASE_NAME=db.sqlite3
   
   # Redis Configuration (for background tasks)
   REDIS_URL=redis://localhost:6379
   
   # Kinopoisk API (for fetching movie data)
   KINOPOISK_API_KEY=your_api_key_here
   KINOPOISK_API_URL=https://kinopoiskapiunofficial.tech/api
   KINOPOISK_API_VERSION=v2.2
   ```
   
   Get your API keys:
   - **Telegram Bot Token**: [@BotFather](https://t.me/botfather)
   - **Group ID**: Forward message from group to [@userinfobot](https://t.me/userinfobot)
   - **Kinopoisk API**: [kinopoisk.dev](https://kinopoiskapiunofficial.tech/)

4. **Initialize database**
   
   Database migrations run automatically on first start via GORM AutoMigrate.
   
   **Tables created**:
   - `users` - Telegram user accounts
   - `roles` - User roles (seeded: ADMIN, USER)
   - `movies` - Movie catalog
   - `sessions` - Viewing sessions
   - `votings` - Voting instances
   - `votes` - Individual vote records
   - `polls` - Telegram poll tracking
   - `poll_options` - Poll option mappings
   - `schedules` - Recurring schedule configuration
   - `movies_sessions` - Many-to-many relationship table
   
   **Optional**: Import existing movies:
   ```bash
   # From JSON
   go run scripts/import_movies/import_movies.go -file movies.json
   ```

## 🎯 Usage

### Running Locally

**Start Redis** (required for background tasks):
```bash
redis-server
```

**Run the bot**:
```bash
# Using Make
make run

# Or directly with Go
go run ./cmd/bot/main.go
```

**Run the worker** (in separate terminal):
```bash
go run ./cmd/worker/main.go
```

**Build and run**:
```bash
# Build both bot and worker
make build

# Run bot
./build/bot

# Run worker (in separate terminal)
./build/worker
```

### Running with Docker

**Development** (with docker-compose):
```bash
# Build and start all services (redis, bot, worker)
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

**Production** (with nginx):
```bash
# Build and start with nginx reverse proxy
docker-compose -f docker-compose.prod.yml up --build -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Stop services
docker-compose -f docker-compose.prod.yml down
```

### Docker Build Targets

The Dockerfile supports building both bot and worker from the same source:

```bash
# Build bot
docker build --build-arg BUILD_TARGET=bot -t movie-club-bot .

# Build worker
docker build --build-arg BUILD_TARGET=worker -t movie-club-worker .
```

### Bot Commands

#### User Commands
- `/start` - Start the bot and get welcome message
- `/help` - Show help message with all available commands
- `/suggest` - Suggest a movie (send Kinopoisk links or IDs)
- `/current` - Show current session movies
- `/watched` - Show already watched movies (paginated Telegraph list)
- `/cancel` - Cancel current operation/conversation flow

#### Admin Commands
- `/adds <movie_ids>` - Add movies to current session
- `/removes` - Remove movies from current session
- `/cancel_session` - Cancel current viewing session
- `/reschedule` - Reschedule current session date/time
- `/custom` - Set custom description for current session
- `/voting` - Create a new voting (selection or rating)
- `/cancel_voting` - Cancel active votings
- `/schedule` - View current schedule
- `/reschedule_schedule` - Update recurring schedule settings

### Workflows

#### Creating a Viewing Session
1. Admin uses `/adds <movie_ids>` with Kinopoisk IDs or links
2. Bot fetches movie information from Kinopoisk API
3. Session is created automatically (or movies added to existing)
4. Bot schedules:
   - Session finish task
   - Rating voting tasks for each movie (opens at session end)

#### Managing Sessions
- **Add Description**: `/custom` - Set custom description with max 500 chars
- **Reschedule**: `/reschedule` - Choose new date, time, and timezone
- **Remove Movies**: `/removes` - Select movies to remove from session
- **Cancel**: `/cancel_session` - Cancel session and all related tasks

#### Creating Votings

1. **Selection Voting** (Choose next movie):
   - Admin runs `/voting` → selects "Selection"
   - Enters title and duration (hours)
   - Selects movies from paginated list
   - Bot creates Telegram poll
   - Poll auto-closes after duration, movie with most votes wins

2. **Rating Voting** (Rate watched movie):
   - Admin runs `/voting` → selects "Rating"
   - Selects watched movies from list
   - Polls are created immediately (or scheduled automatically after session)
   - Members rate 1-10
   - Average rating is calculated and saved

#### Suggesting Movies
1. User sends message with Kinopoisk links or IDs
2. Bot parses links/IDs (supports multiple per message, max 5)
3. Checks if movies already exist
4. Fetches new movie data from Kinopoisk
5. Adds movies to database with status "SUGGESTED"

## 🛠️ Development

### Available Make Commands

```bash
# Build the bot binary
make build

# Run the bot
make run

# Format code
make fmt

# Run go vet
make vet

# Run tests
make test

# Run all checks (fmt, vet, test)
make check

# Clean build artifacts
make clean
```

### Code Structure Guidelines

- **Handlers**: Keep telegram handlers focused on I/O, delegate to services
- **Services**: Implement business logic, coordinate between repositories
- **Repositories**: Keep data access logic isolated, use interfaces
- **Models**: Define GORM entities with proper relationships
- **Middleware**: Add authentication/authorization logic
- **Tasks**: Keep background tasks idempotent and retryable

### Adding New Commands

1. Create handler in `internal/transport/telegram/`
2. Implement service logic in `internal/service/`
3. Add repository methods if needed in `internal/repository/`
4. Register handler in `internal/app/app.go`
5. Add middleware if admin-only
6. Update this README with command documentation

### Adding Background Tasks

1. Create task file in `internal/tasks/`
2. Define payload struct and enqueue function
3. Implement processor function
4. Register in worker's main.go
5. Enqueue from handler/service where needed

## 📊 Database Schema

### Core Tables

- **users**: Telegram users (ID, username, first name, last name)
- **roles**: User roles (ADMIN, USER) - seeded automatically
- **movies**: Movie information
  - ID (Kinopoisk ID), Title, Description, Directors
  - Year, Countries, Genres, Link, Duration
  - IMDBRating, Rating (calculated from votes)
  - Status (SUGGESTED/WATCHED), WatchCount
  - FinishedAt, SuggestedAt, SuggestedBy
- **sessions**: Movie viewing sessions
  - FinishedAt (Unix timestamp)
  - Status (ONGOING/FINISHED/CANCELLED)
  - Description (custom description)
  - CreatedBy (user ID)
- **votings**: Voting sessions
  - Title, Status (ACTIVE/CLOSED/CANCELLED)
  - Type (SELECTION/RATING), CreatedBy
  - SessionID (optional link to session)
- **votes**: Individual votes
  - UserID, VotingID, MovieID
  - Value (for rating votes: 1-10)
- **polls**: Telegram poll tracking (persistence across restarts)
  - PollID (Telegram poll ID)
  - MessageID, ChatID, VotingID
- **poll_options**: Maps poll options to movies/votings
  - PollID, OptionID, MovieID, VotingID
- **schedules**: Recurring schedule configuration
  - Weekday (1-7), Hour, Minute
  - Location (timezone), IsActive, Description

### Relationships

```
users ──→ roles (many-to-one)
users ──→ movies (one-to-many, via SuggestedBy)
users ──→ sessions (one-to-many, via CreatedBy)
users ──→ votings (one-to-many, via CreatedBy)
users ──→ votes (one-to-many)

movies ←→ sessions (many-to-many via movies_sessions)
movies ──→ votes (one-to-many)
movies ──→ poll_options (one-to-many)

sessions ──→ votings (one-to-many)

votings ──→ votes (one-to-many)
votings ──→ polls (one-to-many)

polls ──→ poll_options (one-to-many)
```

## 🔧 Utilities & Scripts

### Export Movies to JSON

Export all movies from database to `movies.json`:
```bash
go run scripts/export_movies/export_movies.go
```

### Import Movies from JSON

Import movies from JSON file to database:
```bash
go run scripts/import_movies/import_movies.go -file movies.json
```

### Generate SQL INSERT Statements

Convert `movies.json` to SQL INSERT statements in `movies_insert.sql`:
```bash
go run scripts/json_to_sql/json_to_sql.go
```

## 🏗️ Architecture

### Layered Architecture

The project follows clean architecture principles:

1. **Transport Layer** (`internal/transport/telegram/`): 
   - Telegram update handlers
   - Command routing
   - Message formatting with emoji
   - Inline keyboard management

2. **Service Layer** (`internal/service/`): 
   - Business logic implementation
   - Data validation
   - Cross-repository coordination
   - External API integration (Kinopoisk, Telegraph)

3. **Repository Layer** (`internal/repository/`): 
   - Data access abstraction
   - GORM database operations
   - Interface-based design for testability

4. **Model Layer** (`internal/model/`): 
   - GORM entity definitions
   - Database schema structure
   - Entity relationships

### State Management (FSM)

Uses Finite State Machine for complex conversation flows:
- **Default State**: Normal command handling
- **Voting Creation Flow**: Type → Title → Duration → Movies selection
- **Session Management**: Date → Time → Location selection
- **Custom Description**: Description input and validation
- **Movie/Voting Removal**: List display → Index input → Confirmation

States tracked per user with data persistence across interactions.

### Background Task System

**Architecture**: Separate bot and worker processes

**Bot Process** (`cmd/bot/`):
- Handles Telegram updates
- Enqueues background tasks
- Inspects task status

**Worker Process** (`cmd/worker/`):
- Processes queued tasks
- Executes scheduled jobs
- Manages task retry logic

**Task Types**:
1. **FinishSession**: Closes session at scheduled time
2. **OpenRatingVoting**: Creates rating polls when session ends
3. **CloseSelectionVoting**: Closes selection voting, determines winner
4. **CloseRatingVoting**: Closes rating poll, calculates average

**Scheduling**:
- Tasks scheduled with `ProcessIn` duration
- Unique task IDs prevent duplicates
- Task inspection for status checking
- Task deletion on session cancellation

### Middleware System

**Permission Middleware** (`utils/telegram/middleware/`):
- `CheckIfInGroup`: Verifies group membership, auto-registers new users
- `CheckIfAdmin`: Verifies admin permissions
- Role detection based on Telegram administrator status
- Automatic role assignment (USER/ADMIN)

### Docker Architecture

**Multi-Stage Build** (`Dockerfile`):
- Build stage: Go 1.25.3-alpine with CGO for Postgres
- Runtime stage: Minimal alpine with ca-certificates
- Build target selection via `BUILD_TARGET` arg (bot/worker)

**Orchestration**:
- `docker-compose.prod.yml`: Production without services

## 🐛 Troubleshooting

### Bot doesn't respond
- Verify bot token in `.env` is correct
- Check group ID is correct (negative number for supergroups)
- Ensure bot is added to the group and has message access
- Check bot logs for errors

### "No permission" errors
- User needs to be registered first (send any message)
- Admin commands require Telegram administrator status
- Check middleware is properly applied

### Background tasks not running
- Ensure Redis is running and accessible
- Verify worker process is started
- Check Redis URL in `.env` is correct
- Review worker logs for task processing

### Database errors
- Ensure SQLite CGO support is enabled during build
- Check file permissions on `db.sqlite3`
- Verify migrations completed (check logs on first start)

### Docker issues
- Ensure ports 6379 (redis) are available
- Check `.env` file is in same directory as docker-compose.yml
- Verify Docker has sufficient resources allocated
- Check logs: `docker-compose logs -f`

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/tg-bot-movieclub-go.git
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make your changes**
   - Follow existing code style
   - Add emoji to user-facing messages
   - Write clear commit messages
   - Update documentation if needed

4. **Test your changes**
   ```bash
   make check
   ```

5. **Commit and push**
   ```bash
   git commit -m '✨ Add amazing feature'
   git push origin feature/amazing-feature
   ```

6. **Open a Pull Request**
   - Describe your changes clearly
   - Reference any related issues
   - Wait for review

### Commit Message Guidelines
- ✨ `:sparkles:` - New feature
- 🐛 `:bug:` - Bug fix
- 📝 `:memo:` - Documentation
- ♻️ `:recycle:` - Refactoring
- 🎨 `:art:` - UI/formatting improvements
- ⚡ `:zap:` - Performance improvements
- 🔧 `:wrench:` - Configuration changes

## 📝 License

This project is licensed under the MIT License.

## 👤 Author

**Forceres**

- GitHub: [@Forceres](https://github.com/Forceres)

## �️ Roadmap

### Current Features (v1.0)
- ✅ Movie suggestions and management
- ✅ Session creation and scheduling
- ✅ Selection and rating votings
- ✅ Background task scheduling
- ✅ Admin/user role system
- ✅ Auto-registration
- ✅ Emoji-enhanced messages
- ✅ Docker deployment

### Under Consideration
- Custom voting types
- Integration with streaming services
- Movie trailer embedding
- Discussion threads per movie
- Voting result analytics

## �🙏 Acknowledgments

This project is built with these amazing open-source libraries:

- **[go-telegram/bot](https://github.com/go-telegram/bot)** - Telegram Bot API wrapper with comprehensive support
- **[GORM](https://gorm.io/)** - Fantastic ORM library for Go with auto-migrations
- **[Asynq](https://github.com/hibiken/asynq)** - Reliable task queue and scheduler built on Redis
- **[go-telegram/fsm](https://github.com/go-telegram/fsm)** - State machine for conversation flows
- **[go-telegram/ui](https://github.com/go-telegram/ui)** - UI components including paginator
- **[Telegraph Go](https://github.com/celestix/telegraph-go)** - Telegraph API client for content publishing
- **[cleanenv](https://github.com/ilyakaznacheev/cleanenv)** - Clean and elegant environment configuration
- **[Kinopoisk Unofficial API](https://kinopoiskapiunofficial.tech/)** - Movie data source

Special thanks to all contributors and the Go community!