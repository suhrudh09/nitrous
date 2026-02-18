# NITROUS Backend — Go API Server

High-performance REST API for the Nitrous motorsports platform.

## Tech Stack

- **Go 1.21+** — Backend language
- **Gin** — HTTP web framework
- **JWT** — Authentication
- **In-memory storage** — For prototype (easily swappable with PostgreSQL)

## Project Structure

```
nitrous-backend/
├── main.go              — Entry point, router setup
├── config/              — Environment configuration
├── database/            — Database connection & seed data
├── models/              — Data models (Event, User, etc.)
├── handlers/            — HTTP request handlers (controllers)
├── middleware/          — Auth middleware, CORS, etc.
└── utils/               — Helper functions (JWT, etc.)
```

## Setup Instructions

### 1. Install Go

```bash
# macOS
brew install go

# Linux
sudo apt install golang-go

# Verify
go version
```

### 2. Clone and Run

```bash
cd nitrous-backend

# Install dependencies
go mod download

# Run server
go run main.go
```

Server starts on **http://localhost:8080**

### 3. Test API

```bash
# Health check
curl http://localhost:8080/health

# Get all events
curl http://localhost:8080/api/events

# Get live events only
curl http://localhost:8080/api/events/live

# Get categories
curl http://localhost:8080/api/categories
```

## API Endpoints

### Public Routes

#### Events
```
GET    /api/events              List all events
GET    /api/events/live         Get only live events
GET    /api/events/:id          Get event by ID
GET    /api/events?category=motorsport  Filter by category
```

#### Categories
```
GET    /api/categories          List all categories
GET    /api/categories/:slug    Get category by slug
```

#### Journeys
```
GET    /api/journeys            List all journeys
GET    /api/journeys/:id        Get journey by ID
```

#### Merch
```
GET    /api/merch               List all merch items
GET    /api/merch/:id           Get merch item by ID
```

#### Teams
```
GET    /api/teams               List all teams
GET    /api/teams/:id           Get team by ID
```

#### Streams
```
GET    /api/streams             List all streams
GET    /api/streams/:id         Get stream by ID
GET    /api/streams/ws          WebSocket endpoint for telemetry stream
```

#### Auth
```
POST   /api/auth/register       Create account
POST   /api/auth/login          Login (returns JWT)
```

### Protected Routes (require JWT)

```
GET    /api/auth/me             Get current user info
POST   /api/events              Create event
PUT    /api/events/:id          Update event
DELETE /api/events/:id          Delete event
POST   /api/journeys/:id/book   Book a journey
POST   /api/teams/:id/follow    Follow a team
POST   /api/teams/:id/unfollow  Unfollow a team
GET    /api/reminders           List my reminders
POST   /api/reminders           Create reminder
DELETE /api/reminders/:id       Delete reminder
POST   /api/orders              Create order
GET    /api/orders              List my orders
GET    /api/orders/:id          Get my order by ID
```

## Authentication

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepass123",
    "name": "John Doe"
  }'
```

Response:
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "createdAt": "2026-02-16T..."
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepass123"
  }'
```

### Use Token
```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Teams (Public + Protected)
```bash
# List teams
curl http://localhost:8080/api/teams

# Get a team by ID
curl http://localhost:8080/api/teams/TEAM_ID

# Follow team (auth required)
curl -X POST http://localhost:8080/api/teams/TEAM_ID/follow \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Unfollow team (auth required)
curl -X POST http://localhost:8080/api/teams/TEAM_ID/unfollow \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Streams (REST + WebSocket)
```bash
# List streams
curl http://localhost:8080/api/streams

# Get stream by ID
curl http://localhost:8080/api/streams/stream-1

# WebSocket endpoint (use a WS client like wscat)
wscat -c ws://localhost:8080/api/streams/ws
```

### Orders (Protected)
```bash
# Create order
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchItemId": "MERCH_ITEM_ID",
    "quantity": 2
  }'

# List my orders
curl http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get my order by ID
curl http://localhost:8080/api/orders/ORDER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Reminders (Protected)
```bash
# Create reminder
curl -X POST http://localhost:8080/api/reminders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "eventId": "EVENT_ID",
    "message": "Notify me 1 hour before",
    "remindAt": "2026-03-15T14:00:00Z"
  }'

# List my reminders
curl http://localhost:8080/api/reminders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Delete reminder
curl -X DELETE http://localhost:8080/api/reminders/REMINDER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## CORS Configuration

Frontend origins are whitelisted:
- `http://localhost:3000` (Next.js dev)
- `https://nitrous.vercel.app` (Production)

## Environment Variables

Create `.env` file:

```env
# Database (for production)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=nitrous

# JWT Secret (change this!)
JWT_SECRET=your-super-secret-key-change-this

# Server Port
PORT=8080
```

## Switching to PostgreSQL

### 1. Install PostgreSQL

```bash
# macOS
brew install postgresql

# Start service
brew services start postgresql
```

### 2. Create Database

```sql
CREATE DATABASE nitrous;
```

### 3. Update database/db.go

Replace in-memory storage with actual DB connection:

```go
package database

import (
    "database/sql"
    "log"
    _ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
    connStr := "host=localhost port=5432 user=postgres password=yourpass dbname=nitrous sslmode=disable"
    
    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    if err = DB.Ping(); err != nil {
        log.Fatal("Database unreachable:", err)
    }
    
    log.Println("✓ Connected to PostgreSQL")
}
```

### 4. Run Migrations

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    date TIMESTAMP,
    time VARCHAR(50),
    is_live BOOLEAN DEFAULT FALSE,
    category VARCHAR(50),
    thumbnail_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Production Deployment

### Railway (Recommended)

1. Install Railway CLI:
```bash
npm i -g @railway/cli
```

2. Deploy:
```bash
railway login
railway init
railway up
```

### Fly.io

1. Install flyctl:
```bash
curl -L https://fly.io/install.sh | sh
```

2. Deploy:
```bash
fly launch
fly deploy
```

## Development Tips

### Hot Reload

Install Air for hot reloading:

```bash
go install github.com/cosmtrek/air@latest
air
```

### Testing

```bash
go test ./...
```

### Build Binary

```bash
go build -o nitrous-api
./nitrous-api
```

## Next Steps

- [ ] Expand live telemetry pipeline (producer integration + persistence)
- [ ] Implement payment gateway (Stripe)
- [ ] Add email verification
- [ ] Rate limiting
- [ ] Logging middleware
- [ ] Unit tests
- [ ] API documentation (Swagger)

## License

MIT
