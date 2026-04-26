# TodoAI - Production-Ready Todo Application

A modern, full-stack todo application built with Go, PostgreSQL, and vanilla HTML/CSS/JavaScript. Designed for production deployment with robust infrastructure, responsive design, and mobile-first approach.

## Features

### Core Functionality
- **User Authentication**: Register/Login with JWT-based authentication
- **CRUD Operations**: Create, read, update, delete todo items
- **Filtering & Sorting**: Filter by status, sort by priority/due date/created
- **Statistics Dashboard**: Real-time stats with completion rate
- **Quick Filters**: Today's tasks, overdue tasks
- **Rich Todo Details**: Priority levels, due dates, notes

### Technical Features
- **Production Infrastructure**: Docker multi-stage builds, health checks
- **Database**: PostgreSQL with migrations, indexes, triggers
- **API**: RESTful JSON API with proper HTTP status codes
- **Middleware**: CORS, rate limiting, request timeout, recovery, logging
- **Security**: Password hashing (bcrypt), JWT tokens, input validation
- **Observability**: Structured logging (Zap), health endpoints
- **Dev Experience**: Hot reload (Air), Makefile commands, comprehensive scripts

## Project Structure

```
├── cmd/server/           # Main application entrypoint
├── internal/
│   ├── config/           # Configuration management
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   └── service/          # Business logic layer
├── pkg/
│   ├── database/         # Database connection & utilities
│   └── logger/           # Structured logging setup
├── migrations/           # Database migrations (SQL files)
├── scripts/              # Utility scripts
├── static/               # Frontend assets
│   ├── index.html
│   ├── css/
│   │   └── style.css     # Responsive mobile-first CSS
│   └── js/
│       ├── auth.js       # Authentication logic
│       ├── todos.js      # Todo operations
│       └── app.js        # Main app orchestration
├── configs/              # Configuration files
├── docs/                 # Documentation
├── Dockerfile            # Production Docker image
├── docker-compose.yml    # Development environment
├── docker-compose.prod.yml # Production deployment
├── Makefile              # Development commands
├── .env.example          # Environment template
├── .env                  # Local environment (gitignored)
├── .air.toml             # Hot reload config
└── go.mod                # Go dependencies
```

## Quick Start (Development)

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (optional, for native builds)
- GNU Make (optional)

### One-Command Setup

```bash
# Clone and setup
git clone <your-repo>
cd todo-ai

# Development environment (Docker)
make dev

# Or manual setup
docker-compose up -d
```

### Access the Application

- **Frontend**: http://localhost:8080
- **API**: http://localhost:8080/api
- **Health Check**: http://localhost:8080/health
- **Database**: localhost:5432 (user: `todo_user`, password: `todo_pass`)

### Manual Setup (without Docker)

```bash
# Install dependencies
go mod download

# Setup environment
cp .env.example .env

# Run database migrations
./scripts/migrate.sh up

# Build and run
make build
./bin/server
```

### Development Commands

```bash
make help              # Show all available commands
make dev               # Start dev environment with Docker
make build             # Build binary
make test              # Run tests with coverage
make lint              # Run linter
make fmt               # Format code
make clean             # Clean build artifacts
make migrate-up        # Run database migrations
make migrate-down      # Rollback migrations
make logs              # View application logs
make db-shell          # Connect to database
```

## Production Deployment

### Using Docker Compose

```bash
# 1. Set production environment variables
cp .env.example .env
# Edit .env with production values (DB_PASSWORD, JWT_SECRET, etc.)

# 2. Set up secrets (recommended)
mkdir -p secrets
echo "your-production-db-password" > secrets/db_password.txt

# 3. Use production compose file
docker-compose -f docker-compose.prod.yml up -d

# 4. Optional: Build and push to registry
export IMAGE_TAG=your-registry/todo-app:v1.0.0
make docker-build
make docker-push
```

### Using Kubernetes (Optional)

```yaml
# Example deployment.yaml (TODO)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: todo-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: todo
  template:
    metadata:
      labels:
        app: todo
    spec:
      containers:
      - name: todo-app
        image: your-registry/todo-app:latest
        ports:
        - containerPort: 8080
        envFrom:
        - secretRef:
            name: todo-secrets
        - configMapRef:
            name: todo-config
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `APP_PORT` | HTTP server port | 8080 |
| `APP_DEBUG` | Enable debug mode | false |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | Database user | todo_user |
| `DB_PASSWORD` | Database password | todo_pass |
| `DB_NAME` | Database name | todo_db |
| `DB_SSLMODE` | SSL mode for DB | disable |
| `DB_MAX_OPEN_CONNS` | Max open connections | 25 |
| `DB_MAX_IDLE_CONNS` | Max idle connections | 5 |
| `DB_MAX_LIFETIME` | Connection lifetime | 5m |
| `CORS_ALLOW_ORIGINS` | Allowed CORS origins | * |
| `JWT_SECRET` | JWT signing secret (CHANGE IN PROD!) | dev-key |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | info |
| `LOG_FORMAT` | Log format (json/console) | json |

**Security Note**: Always change `JWT_SECRET` and database credentials in production!

## API Documentation

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | User login |
| GET | `/api/auth/me` | Get current user |
| POST | `/api/auth/refresh` | Refresh JWT token |

### Todo Endpoints (Require Auth)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/todos` | List todos (query: ?completed, ?limit, ?offset) |
| POST | `/api/todos` | Create new todo |
| GET | `/api/todos/{id}` | Get todo by ID |
| PUT | `/api/todos/{id}` | Update todo |
| DELETE | `/api/todos/{id}` | Delete todo |
| GET | `/api/todos/today` | Get today's todos |
| GET | `/api/todos/overdue` | Get overdue todos |
| GET | `/api/todos/stats` | Get statistics |

### Example Request

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","password":"secret123"}'

# Create todo (with token)
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"title":"Buy groceries","priority":2,"notes":"Milk, eggs, bread"}'
```

## Database Schema

### Tables

**users**
- `id` (UUID, PK)
- `username` (VARCHAR(50), UNIQUE)
- `email` (VARCHAR(255), UNIQUE)
- `password_hash` (VARCHAR(255))
- `created_at`, `updated_at` (timestamps)

**todos**
- `id` (UUID, PK)
- `user_id` (UUID, FK → users)
- `title` (VARCHAR(255))
- `completed` (BOOLEAN)
- `priority` (INTEGER 1-5)
- `due_date` (TIMESTAMP)
- `notes` (TEXT)
- `created_at`, `updated_at` (timestamps)

### Indexes
- `idx_todos_user_id` - Fast lookup by user
- `idx_todos_completed` - Filter by completion status
- `idx_todos_due_date` - Sort/filter by due date

### Triggers
- Automatic `updated_at` timestamp updates

## Testing

```bash
# Unit tests
go test -v ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests (requires Docker)
make test-integration
```

## Performance & Monitoring

- **Health Checks**: `/health` (app + DB connectivity), `/readiness`
- **Structured Logging**: JSON logs for easy parsing (ELK/Loki)
- **Metrics**: Application metrics via Prometheus (TODO)
- **Rate Limiting**: 100 requests/second per IP
- **Request Timeout**: 30s timeout enforced
- **Connection Pooling**: Configurable DB pool settings

## Security Considerations

- Passwords hashed with bcrypt (cost 10)
- JWT tokens with 24h expiry
- CORS with configurable allowed origins
- SQL parameterized queries (no SQL injection)
- Rate limiting to prevent abuse
- Request timeouts to prevent DoS
- Input validation on all endpoints

## Contributing

1. Fork the repository
2. Create a feature branch
3. Follow Go conventions and coding standards
4. Add tests for new functionality
5. Run `make all` before committing
6. Submit a pull request

## License

MIT

## Support

For issues and questions, please open an issue on GitHub.