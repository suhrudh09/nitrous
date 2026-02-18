# NITROUS — Full-Stack Architecture Plan
## Prototype 1 Submission Strategy

## Tech Stack Decision

### Frontend: Next.js (TypeScript)
✓ React-based (component reusability)
✓ TypeScript (type safety)
✓ Server-side rendering (fast initial load)
✓ API routes (can proxy to Go backend)
✓ Production-ready deployment (Vercel)

### Backend: Go
✓ High performance (concurrency)
✓ Statically typed (reliable)
✓ Great for APIs (gorilla/mux, gin, fiber)
✓ Easy deployment (single binary)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                    FRONTEND                          │
│              Next.js + TypeScript                    │
│                                                      │
│  ┌──────────────────────────────────────────┐      │
│  │ Static Pages (SSR)                       │      │
│  │  - Hero, Categories, Events List         │      │
│  │  - Pre-rendered for speed                │      │
│  └──────────────────────────────────────────┘      │
│                      ↓                               │
│  ┌──────────────────────────────────────────┐      │
│  │ Interactive Components (Client)          │      │
│  │  - Search/Filter                         │      │
│  │  - Live Stream Player                    │      │
│  │  - Real-time Event Updates               │      │
│  └──────────────────────────────────────────┘      │
│                      ↓                               │
│         API Calls (fetch/axios)                     │
└──────────────────────┬──────────────────────────────┘
                       │
                       │ HTTP/WebSocket
                       ↓
┌─────────────────────────────────────────────────────┐
│                    BACKEND                           │
│                   Go + Gin                           │
│                                                      │
│  ┌──────────────────────────────────────────┐      │
│  │ REST API Endpoints                       │      │
│  │  GET  /api/events                        │      │
│  │  GET  /api/events/:id                    │      │
│  │  POST /api/auth/login                    │      │
│  │  GET  /api/categories                    │      │
│  └──────────────────────────────────────────┘      │
│                      ↓                               │
│  ┌──────────────────────────────────────────┐      │
│  │ Business Logic                           │      │
│  │  - Event filtering                       │      │
│  │  - User authentication                   │      │
│  │  - Payment processing                    │      │
│  └──────────────────────────────────────────┘      │
│                      ↓                               │
│  ┌──────────────────────────────────────────┐      │
│  │ Database Layer (PostgreSQL)              │      │
│  │  - Events, Users, Transactions           │      │
│  └──────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────┘
```

---

## Prototype 1 Scope

### What We'll Build

#### Frontend (Next.js)
- [x] Full UI with all sections
- [x] Responsive design
- [x] Animations (Arise bounce, Tron aesthetic)
- [ ] Connect to Go backend API
- [ ] Real-time updates (WebSocket for live events)
- [ ] Search/filter functionality
- [ ] Modal interactions

#### Backend (Go)
- [x] REST API server (Gin framework)
- [ ] Database schema (PostgreSQL)
- [x] CRUD endpoints for events
- [x] Authentication (JWT)
- [x] CORS setup (allow Next.js origin)
- [x] WebSocket for live data (streams telemetry hub)
- [x] Teams endpoints (list, detail, follow/unfollow)
- [x] Orders endpoints (create, list mine, detail mine)
- [x] Reminders endpoints (create, delete, list mine)

#### Database Schema
```sql
-- Events table
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

-- Categories table
CREATE TABLE categories (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) UNIQUE,
  icon VARCHAR(10),
  live_count INTEGER DEFAULT 0,
  description TEXT,
  color VARCHAR(50)
);

-- Users table (for auth)
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name VARCHAR(255),
  created_at TIMESTAMP DEFAULT NOW()
);
```

---

## Development Workflow

### Phase 1: Backend Foundation (Go)
1. Initialize Go project
2. Set up Gin router
3. Create database models
4. Build REST endpoints
5. Test with Postman/curl

### Phase 2: Frontend Integration
1. Update Next.js to fetch from Go API
2. Replace mock data with real API calls
3. Add loading states
4. Error handling

### Phase 3: Real-time Features
1. WebSocket server in Go
2. Live event updates
3. Real-time status indicators

### Phase 4: Polish
1. Authentication flow
2. Form validation
3. Performance optimization
4. Deployment

---

## API Endpoints Design

### Events
```
GET    /api/events              - List all events
GET    /api/events/:id          - Get single event
GET    /api/events/live         - Get only live events
POST   /api/events              - Create event (admin)
PUT    /api/events/:id          - Update event (admin)
DELETE /api/events/:id          - Delete event (admin)
```

### Categories
```
GET    /api/categories          - List all categories
GET    /api/categories/:slug    - Get category by slug
```

### Auth
```
POST   /api/auth/register       - Create account
POST   /api/auth/login          - Login (returns JWT)
POST   /api/auth/logout         - Logout
GET    /api/auth/me             - Get current user
```

### Teams
```
GET    /api/teams               - List teams
GET    /api/teams/:id           - Get team by ID
POST   /api/teams/:id/follow    - Follow a team (auth required)
POST   /api/teams/:id/unfollow  - Unfollow a team (auth required)
```

### Streams
```
GET    /api/streams             - List streams
GET    /api/streams/:id         - Get stream by ID
GET    /api/streams/ws          - WebSocket telemetry stream
```

### Reminders
```
GET    /api/reminders           - List current user's reminders
POST   /api/reminders           - Create reminder
DELETE /api/reminders/:id       - Delete reminder
```

### Journeys
```
GET    /api/journeys            - List journeys
POST   /api/journeys/:id/book   - Book a journey
```

### Merch
```
GET    /api/merch               - List merch items
POST   /api/orders              - Create order
GET    /api/orders              - List current user's orders
GET    /api/orders/:id          - Get current user's order by ID
```

---

## Folder Structure

### Backend (Go)
```
nitrous-backend/
├── main.go                 ← Entry point
├── go.mod                  ← Dependencies
├── config/
│   └── config.go          ← DB connection, env vars
├── models/
│   ├── event.go           ← Event struct
│   ├── category.go        ← Category struct
│   └── user.go            ← User struct
├── handlers/
│   ├── events.go          ← Event endpoints
│   ├── categories.go      ← Category endpoints
│   ├── auth.go            ← Auth endpoints
│   ├── teams.go           ← Team endpoints
│   ├── streams.go         ← Stream + websocket endpoints
│   ├── orders.go          ← Order endpoints
│   └── reminders.go       ← Reminder endpoints
├── middleware/
│   ├── auth.go            ← JWT validation
│   └── cors.go            ← CORS config
├── database/
│   └── db.go              ← Database setup
└── utils/
    └── response.go        ← JSON helpers
```

### Frontend (Next.js)
```
nitrous-app/
├── app/                    ← Routes
├── components/             ← UI components
├── lib/
│   ├── api.ts             ← API client (fetch wrapper)
│   └── types.ts           ← Shared types
└── hooks/
    └── useEvents.ts       ← Custom hook for fetching events
```

---

## Deployment Strategy

### Development
- **Frontend:** `npm run dev` on localhost:3000
- **Backend:** `go run main.go` on localhost:8080
- **Database:** PostgreSQL on localhost:5432 (or Docker)

### Production
- **Frontend:** Deploy to Vercel (automatic from GitHub)
- **Backend:** Deploy to Railway/Fly.io/AWS EC2
- **Database:** Supabase/Railway Postgres/AWS RDS

---

## Timeline Estimate

### Prototype 1 (MVP)
- **Week 1:** Backend API + Database
- **Week 2:** Frontend integration + Real-time
- **Week 3:** Auth + Polish
- **Week 4:** Testing + Deployment

### For Rapid Submission
- **Day 1-2:** Backend skeleton + API endpoints
- **Day 3-4:** Frontend connects to backend
- **Day 5:** Demo-ready prototype

---

## Communication Between Frontend & Backend

### Example: Fetching Events

**Go Backend (handlers/events.go):**
```go
func GetEvents(c *gin.Context) {
    var events []models.Event
    database.DB.Find(&events)
    
    c.JSON(200, gin.H{
        "events": events,
        "count": len(events),
    })
}
```

**Next.js Frontend (lib/api.ts):**
```typescript
export async function getEvents(): Promise<Event[]> {
    const res = await fetch('http://localhost:8080/api/events')
    const data = await res.json()
    return data.events
}
```

**Next.js Page (app/page.tsx):**
```typescript
export default async function Home() {
    const events = await getEvents()
    return <EventList events={events} />
}
```

---

## Key Decisions

### Why Next.js over Angular?
- Faster development (React ecosystem)
- Better SEO (server-side rendering)
- Simpler state management
- Huge community (more resources)
- TypeScript support (same as Angular)

### Why Gin for Go?
- Most popular Go web framework
- Fast routing
- Middleware support (auth, CORS, logging)
- Good documentation

### Why PostgreSQL?
- Relational data (events, users, bookings)
- JSON support (flexible metadata)
- Battle-tested (Uber, Instagram use it)
- Free tier on Railway/Supabase

---

## Next Steps

1. **Decide:** Do we build backend first or integrate as we go?
2. **Setup:** Initialize Go project structure
3. **Build:** Create first API endpoint
4. **Connect:** Link Next.js to Go backend
5. **Iterate:** Add features incrementally

Ready to start?
