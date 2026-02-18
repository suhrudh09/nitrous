# NITROUS — Full-Stack Motorsports Platform
## Prototype 1: Complete Reactive Application

A high-performance motorsports streaming platform with Go backend and Next.js frontend.

---

## 🚀 Quick Start

### Prerequisites
- **Node.js** 18+ (for frontend)
- **Go** 1.21+ (for backend)
- **Git**

### 1. Clone the Repository

```bash
# Extract the project
tar -xzf nitrous-app.tar.gz
cd nitrous-app
```

### 2. Start the Backend (Go)

```bash
# Terminal 1: Go API Server
cd nitrous-backend

# Install dependencies
go mod download

# Run server
go run main.go
```

Backend will start on **http://localhost:8080**

Test it:
```bash
curl http://localhost:8080/health
# Should return: {"status":"ok","message":"Nitrous API is running"}
```

### 3. Start the Frontend (Next.js)

```bash
# Terminal 2: Next.js Frontend
cd nitrous-app

# Install dependencies
npm install

# Run dev server
npm run dev
```

Frontend will start on **http://localhost:3000**

### 4. Open in Browser

Visit **http://localhost:3000**

You should see:
- ✅ Animated hero section with Tron aesthetic
- ✅ Floating circuit traces
- ✅ Bouncing hero nav cards
- ✅ Real data from Go backend API

---

## 📁 Project Structure

```
.
├── nitrous-backend/          Go API Server
│   ├── main.go              Entry point
│   ├── config/              Configuration
│   ├── database/            Data layer (in-memory for prototype)
│   ├── handlers/            API endpoints
│   ├── middleware/          Auth, CORS
│   ├── models/              Data models
│   └── utils/               JWT helpers
│
└── nitrous-app/             Next.js Frontend
    ├── app/                 Routes & pages
    │   ├── layout.tsx       Root layout
    │   ├── page.tsx         Homepage
    │   └── globals.css      Global styles
    ├── components/          React components
    │   ├── Nav.tsx          Navigation
    │   └── Hero.tsx         Hero section
    ├── lib/                 Utilities
    │   ├── api.ts           API client
    │   └── data.ts          Mock data (backup)
    ├── types/               TypeScript definitions
    └── public/              Static assets
```

---

## 🔌 API Endpoints

### Events
```
GET    /api/events              All events
GET    /api/events/live         Live events only
GET    /api/events/:id          Single event
GET    /api/events?category=motorsport  Filter by category
```

### Categories
```
GET    /api/categories          All categories
GET    /api/categories/:slug    Single category
```

### Journeys
```
GET    /api/journeys            All journeys
POST   /api/journeys/:id/book   Book journey (auth required)
```

### Merch
```
GET    /api/merch               All merch items
```

### Teams
```
GET    /api/teams               All teams
GET    /api/teams/:id           Single team
POST   /api/teams/:id/follow    Follow team (auth required)
POST   /api/teams/:id/unfollow  Unfollow team (auth required)
```

### Streams
```
GET    /api/streams             All streams
GET    /api/streams/:id         Single stream
GET    /api/streams/ws          WebSocket telemetry stream
```

### Orders
```
POST   /api/orders              Create merch order (auth required)
GET    /api/orders              Current user's orders (auth required)
GET    /api/orders/:id          Current user's order by ID (auth required)
```

### Reminders
```
GET    /api/reminders           Current user's reminders (auth required)
POST   /api/reminders           Create reminder (auth required)
DELETE /api/reminders/:id       Delete reminder (auth required)
```

### Auth
```
POST   /api/auth/register       Create account
POST   /api/auth/login          Login
GET    /api/auth/me             Current user (auth required)
```

---

## 🧪 Testing the API

### Get Events
```bash
curl http://localhost:8080/api/events
```

### Register User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepass123",
    "name": "Test User"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepass123"
  }'
```

### Teams
```bash
curl http://localhost:8080/api/teams

curl -X POST http://localhost:8080/api/teams/TEAM_ID/follow \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Streams
```bash
curl http://localhost:8080/api/streams

wscat -c ws://localhost:8080/api/streams/ws
```

### Orders
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchItemId": "MERCH_ITEM_ID",
    "quantity": 1
  }'

curl http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Reminders
```bash
curl -X POST http://localhost:8080/api/reminders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "eventId": "EVENT_ID",
    "message": "Ping me before race",
    "remindAt": "2026-03-20T09:00:00Z"
  }'

curl http://localhost:8080/api/reminders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🎨 Features Implemented

### Frontend (Next.js + TypeScript)
- ✅ Server-side rendering for fast initial load
- ✅ TypeScript for type safety
- ✅ CSS Modules for scoped styles
- ✅ Responsive design (mobile, tablet, desktop)
- ✅ Tron-style grid background
- ✅ Animated circuit traces and energy swirls
- ✅ Arise-style bounce animations on hover
- ✅ Floating idle animations
- ✅ Hero nav cards with progress bars
- ✅ API integration with error handling
- ✅ Image optimization (Next/Image)

### Backend (Go + Gin)
- ✅ RESTful API with Gin framework
- ✅ In-memory database (easily swappable with PostgreSQL)
- ✅ JWT authentication
- ✅ Password hashing (bcrypt)
- ✅ CORS configuration
- ✅ Error handling
- ✅ Structured logging
- ✅ Seed data for prototyping
- ✅ Middleware architecture
- ✅ Teams APIs (follow/unfollow)
- ✅ Streams APIs + WebSocket telemetry hub
- ✅ Orders APIs (create + per-user retrieval)
- ✅ Reminders APIs (create/delete/list)

---

## 🔧 Development Tips

### Hot Reload

**Backend (Go):**
```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd nitrous-backend
air
```

**Frontend (Next.js):**
Hot reload is automatic — just save files and see changes instantly.

### Environment Variables

**Backend (.env):**
```env
DB_HOST=localhost
DB_PORT=5432
JWT_SECRET=your-secret-key
PORT=8080
```

**Frontend (.env.local):**
```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

---

## 📦 Production Deployment

### Frontend (Vercel)

1. Push code to GitHub
2. Connect to Vercel
3. Add environment variable:
   ```
   NEXT_PUBLIC_API_URL=https://your-backend-url.com/api
   ```
4. Deploy automatically

### Backend (Railway)

1. Install Railway CLI:
   ```bash
   npm i -g @railway/cli
   ```

2. Deploy:
   ```bash
   cd nitrous-backend
   railway login
   railway init
   railway up
   ```

3. Set environment variables in Railway dashboard

---

## 🗄️ Switching to PostgreSQL

### 1. Install PostgreSQL

```bash
# macOS
brew install postgresql
brew services start postgresql

# Create database
createdb nitrous
```

### 2. Update database/db.go

Replace in-memory storage:

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func InitDB() {
    connStr := "postgres://user:pass@localhost/nitrous?sslmode=disable"
    DB, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("✓ Connected to PostgreSQL")
}
```

### 3. Run Migrations

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    date TIMESTAMP,
    is_live BOOLEAN DEFAULT FALSE,
    category VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 🎯 Next Steps

### Phase 1: Complete UI
- [ ] Categories section with cards
- [ ] Events list with filters
- [ ] Journey cards
- [ ] Merch grid
- [ ] Footer

### Phase 2: Interactivity
- [ ] Search bar (filter events)
- [ ] Live event status updates
- [ ] Modal for "Watch Live" button
- [ ] Booking flow

### Phase 3: Production Features
- [ ] Frontend integration for stream telemetry WebSocket
- [ ] Payment integration (Stripe)
- [ ] Email notifications
- [ ] Admin dashboard
- [ ] Analytics

---

## 📚 Learning Resources

- **Next.js:** https://nextjs.org/docs
- **Go:** https://go.dev/tour/
- **Gin:** https://gin-gonic.com/docs/
- **TypeScript:** https://www.typescriptlang.org/docs

---

## 🐛 Troubleshooting

### Frontend can't connect to backend?
- Check backend is running on port 8080
- Check `.env.local` has correct API URL
- Check browser console for CORS errors

### Backend won't start?
- Check port 8080 is free: `lsof -i :8080`
- Check Go is installed: `go version`
- Check dependencies: `go mod download`

### Styles not loading?
- Restart Next.js dev server
- Clear browser cache
- Check CSS Module imports

---

## 📄 License

MIT

---

## 👥 Contributors

Anurag Achanta ( Front-end Dev-Sec-Ops )
Rishikesh Nalla (Front-end Dev )
Suhrudh Reddy Pyata ( Back-end )
Radhey Sharma ( Back-end )

**Tech Stack:**
- Frontend: Next.js 14 + TypeScript + CSS Modules
- Backend: Go 1.21 + Gin Framework
- Database: In-memory (PostgreSQL-ready)
- Auth: JWT
- Deployment: Vercel + Railway

🚀 **Status:** Prototype 1 Complete — Ready for Demo
