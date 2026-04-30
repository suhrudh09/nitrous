# Sprint 4

## Running the Application

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+ (optional — backend falls back to in-memory store)

### Backend

```bash
cd nitrous-backend

# (Optional) Apply database schema and seed data to a local PostgreSQL instance:
# psql -U postgres -d nitrous -f database/schema.sql
# psql -U postgres -d nitrous -f database/seed.sql

# Set environment variables (copy and adjust as needed):
# export DB_HOST=localhost
# export DB_PORT=5432
# export DB_USER=postgres
# export DB_PASSWORD=yourpassword
# export DB_NAME=nitrous
# export JWT_SECRET=your_jwt_secret
# export YOUTUBE_API_KEY=your_youtube_api_key   # optional, for Live page search

go run .
# Server starts at http://localhost:8080
```

### Frontend

```bash
cd nitrous-app
npm install
npm run dev
# App available at http://localhost:3000
```

### Running Tests

```bash
# Backend unit tests
cd nitrous-backend
go test ./...

# Frontend unit tests (Jest)
cd nitrous-app
npm run test

# Frontend Cypress E2E tests (requires app running at localhost:3000)
cd nitrous-app
npx cypress run          # headless
npx cypress open         # interactive

# Frontend Cypress component tests
cd nitrous-app
npx cypress run --component
```

---

## Sprint 4 Work Completed

### 1. Backend — PostgreSQL migration & DB work
- Migrated backend database flow to PostgreSQL with schema migrations and a postgres-first seed flow with in-memory fallback.
- Files added/updated: `nitrous-backend/database/db.go`, `nitrous-backend/database/schema.sql` (added), `nitrous-backend/database/seed.sql` (added), updated `nitrous-backend/go.mod` and `nitrous-backend/go.sum`.
- Updated all handlers to use the new DB layer: `auth.go`, `events.go`, `orders.go`, `other.go`, `passes.go`, `reminders.go`, `teams.go`, and new handler files for team/cart/payments/notifications/garage configs.

### 2. Backend — Access passes & journey bookings
- Added pass catalog endpoint (`GET /api/passes/catalog`) returning all available passes with remaining spots.
- Added pass purchase endpoint (`POST /api/passes/:id/purchase`) with quantity support and spot-decrement logic.
- Added journey booking endpoint (`POST /api/journeys/:id/book`) with quantity support.
- Added `GET /api/journeys/my` for authenticated users to list their bookings.
- Files: `nitrous-backend/handlers/passes.go`, `nitrous-backend/handlers/other.go`.

### 3. Backend — Payment flow
- Implemented simulated payment flow: `POST /api/payments/intent` → `POST /api/payments/confirm`.
- Added `GET /api/payments/:id` and `GET /api/payments/my` for payment status and history.
- File: `nitrous-backend/handlers/payments.go`.

### 4. Backend — Cart
- Added authenticated cart endpoints: `GET /api/cart`, `POST /api/cart`, `DELETE /api/cart`.
- Cart saves per-user with dedup validation; syncs guest cart to authenticated user on login.
- File: `nitrous-backend/handlers/cart.go`.

### 5. Backend — Notifications
- Added `GET /api/notifications` and `POST /api/notifications/:id/read` for in-app notifications.
- Background worker (`reminder_notifier.go`) converts due reminders into notifications automatically.
- Files: `nitrous-backend/handlers/notifications.go`, `nitrous-backend/handlers/reminder_notifier.go`.

### 6. Backend — Garage configs
- Added authenticated garage config save/list/delete endpoints.
- File: `nitrous-backend/handlers/garage_configs.go`.

### 7. Backend — Role-based access and plan gating
- Register endpoint forces `role=viewer` and `plan=FREE` on all new signups.
- Added `PUT /api/auth/me/plan` and `PUT /api/auth/me/role` with plan-hierarchy enforcement.
- Added team manager and team relation (member/sponsor) CRUD with role guards.
- Files: `nitrous-backend/handlers/auth.go`, `team_managers.go`, `team_relations.go`, `middleware/auth.go`, `models/models.go`.

### 8. Backend — CORS and YouTube embed
- Updated CORS to allow the Vercel frontend origin.
- Added YouTube video search backend proxy for the Live page.
- Files: `nitrous-backend/main.go`, `nitrous-backend/handlers/youtube.go`, `nitrous-backend/config/config.go`.

### 9. Backend — Comprehensive unit tests (Sprint 4 addition)
- Added `additional_handlers_test.go` covering 10 new test functions for previously untested handlers: plan/role updates, cart, payments, notifications, garage configs, team managers, team relations, pass catalog, garage supplemental endpoints, streams/YouTube/admin.
- Fixed `orders_reminders_test.go` fixture to use `pending` status and `CreatedAt: time.Now()`.
- All tests pass: `go test ./...`.

### 10. Frontend — Payment and plan upgrade flow
- Passes page (`/passes`) is now API-driven with a payment modal and quantity selector; spotsLeft decrements immediately in UI on purchase.
- Journeys page (`/journeys`) is now API-driven with a payment modal and quantity booking.
- Settings page (`/settings`) auto-opens VIP/PLATINUM checkout modal based on `?plan=` query param set by signup redirect; calls `updateCurrentUserPlan` then `updateCurrentUserRole` on payment success.

### 11. Frontend — Signup role-gating
- Login/signup page stores selected role; after signup, viewer lands on home, participant/manager are redirected to settings with VIP checkout pre-opened, sponsor is redirected to settings with PLATINUM checkout pre-opened.
- Role is promoted only after successful payment.
- Files: `nitrous-app/app/login/page.tsx`, `nitrous-app/app/settings/page.tsx`.

### 12. Frontend — Cart and orders
- Cart persists to backend for authenticated users; guest → logged-in cart sync on login.
- Orders page has tabs for merch orders and journey bookings.
- Order detail page has a "Repeat" button that re-adds items to cart.
- Files: `nitrous-app/app/cart/page.tsx`, `nitrous-app/app/orders/page.tsx`, `nitrous-app/app/orders/[id]/page.tsx`.

### 13. Frontend — Garage premium gating
- Guest and viewer users are blocked from saving garage configs; a popup displays "You have discovered a premium feature. Upgrade Now" with a CTA to `/settings` for viewers or to signup for guests.
- File: `nitrous-app/app/garage/page.tsx`.

### 14. Frontend — Reminders page
- Reminders list page with delete and 15-second auto-refresh.
- Files: `nitrous-app/app/reminders/page.tsx`, `nitrous-app/app/reminders/reminders.module.css`.

### 15. Frontend — Tooling and Cypress tests
- Upgraded ESLint and resolved `next-config` dependency conflict.
- Added Cypress E2E and component test suite covering home page, hero interactions, `Nav` component, and `Hero` component.
- Files: `nitrous-app/cypress/e2e/`, `nitrous-app/cypress/component/`.

---

## Frontend Unit Tests

### Jest Unit Tests

**File:** `nitrous-app/__tests__/api.test.ts`

| Suite | Test |
|---|---|
| `getEvents` | fetches events successfully |
| `getEvents` | returns empty array on error |
| `getCategories` | fetches categories successfully |
| `getJourneys` | fetches journeys successfully |
| `getMerchItems` | fetches merch items successfully |
| `getEventById` | fetches event by ID successfully |
| `Authentication Functions` | registers user successfully |
| `Authentication Functions` | logs in user successfully |
| `Authentication Functions` | gets current user with token |
| `Error Handling` | handles API errors properly |
| `Error Handling` | handles network errors |

Run:

```bash
cd nitrous-app
npm run test
```

---

### Cypress Component Tests

**File:** `nitrous-app/cypress/component/Hero.cy.tsx`

| Suite | Test |
|---|---|
| Structure | renders hero section |
| Structure | renders background image wrapper |
| Structure | renders circuit layer |
| Structure | renders hero content container |
| Structure | renders hero nav rail |
| Background and Visual Elements | displays the background image |
| Background and Visual Elements | renders circuit layer traces |
| Background and Visual Elements | renders circuit layer nodes |
| Background and Visual Elements | renders energy swirl animations |
| Background and Visual Elements | renders HUD corner elements |
| HUD and Text Content | displays HUD label with system status |
| HUD and Text Content | displays event qualify text in HUD |
| HUD and Text Content | renders HUD line element |
| HUD and Text Content | renders HUD dot indicator |
| Hero Title and Subtitle | displays main title with NITROUS |
| Hero Title and Subtitle | displays FUEL with glow styling |

**File:** `nitrous-app/cypress/component/Nav.cy.tsx`

| Suite | Test |
|---|---|
| Logged out | renders the NITROUS logo |
| Logged out | logo links to / |
| Logged out | shows nav links: Live, Events, Teams, Journeys, Merch |
| Logged out | nav links have correct hrefs |
| Logged out | shows 4 Events Live status badge |
| Logged out | shows Sign In button when logged out |
| Logged out | does not show user avatar when logged out |
| Logged in | shows user initials avatar instead of Sign In |
| Logged in | renders user menu button with aria-label |
| Logged in | opens dropdown on avatar click |
| Logged in | dropdown contains all menu items |
| Logged in | dropdown menu items have correct hrefs |
| Logged in | closes dropdown on outside click |
| Logged in | closes dropdown when a menu item is clicked |
| Logged in | sign out clears user and shows Sign In |
| Structure | renders a nav element |
| Structure | renders nav center links container |

---

### Cypress E2E Tests

**File:** `nitrous-app/cypress/e2e/home.cy.ts`

| Suite | Test |
|---|---|
| Home Page Navigation | loads the home page successfully |
| Home Page Navigation | displays navigation menu with all links |
| Home Page Navigation | displays Sign In link |
| Home Page Navigation | displays hero title and subtitle |
| Home Page Navigation | displays action buttons |

**File:** `nitrous-app/cypress/e2e/hero-interaction.cy.ts`

| Suite | Test |
|---|---|
| Hero Section Interactions | displays hero action buttons and verifies they are clickable |
| Hero Section Interactions | verifies all navigation cards are visible in hero section |
| Hero Section Interactions | can click the Ignite Stream button |
| Hero Section Interactions | can click the Explore Events button |
| Hero Section Interactions | can navigate through hero nav cards |
| Hero Section Interactions | can navigate to live streams from hero |
| Hero Section Interactions | verifies hero section styling elements exist |

Run E2E (requires app running at `http://localhost:3000`):

```bash
cd nitrous-app
npx cypress run          # headless
npx cypress open         # interactive UI
```

Run component tests:

```bash
cd nitrous-app
npx cypress run --component
```

---

## Backend Unit Tests

### Test Files

| File | Package |
|---|---|
| `handlers/auth_handlers_test.go` | handlers |
| `handlers/admin_management_test.go` | handlers |
| `handlers/additional_handlers_test.go` | handlers |
| `handlers/events_mutations_test.go` | handlers |
| `handlers/garage_passes_test.go` | handlers |
| `handlers/handlers_test.go` | handlers |
| `handlers/journeys_teams_test.go` | handlers |
| `handlers/orders_reminders_test.go` | handlers |
| `handlers/test_helpers_test.go` | handlers |
| `middleware/auth_test.go` | middleware |
| `middleware/admin_test.go` | middleware |
| `utils/jwt_test.go` | utils |

### Test Functions

#### Auth (`auth_handlers_test.go`)
| Function | Description |
|---|---|
| `TestRegisterFlow` | Register returns 200 and sets viewer role + FREE plan |
| `TestLoginFlow` | Login returns JWT token on valid credentials |
| `TestGetCurrentUserFlow` | `/auth/me` returns user data for authenticated request |

#### JWT (`utils/jwt_test.go`)
| Function | Description |
|---|---|
| `TestJWTUtility` | Token generation, parsing, and expiry validation |

#### Middleware (`middleware/`)
| Function | Description |
|---|---|
| `TestAuthMiddleware` | Rejects missing/invalid tokens; passes valid JWT |
| `TestAdminMiddleware` | Allows admin role; blocks non-admin |

#### Read Endpoints (`handlers_test.go`)
| Function | Description |
|---|---|
| `TestGetEvents_ListAndCategoryFilter` | List all events and filter by category query param |
| `TestGetLiveEvents_ReturnsOnlyLive` | `/events/live` returns only live-status events |
| `TestGetEventByID_FoundAndNotFound` | 200 on known ID; 404 on unknown ID |
| `TestCategories_ListAndBySlug` | List categories and fetch by slug |
| `TestJourneys_ListAndByID` | List journeys and fetch by ID |
| `TestMerch_ListAndByID` | List merch items and fetch by ID |
| `TestTeams_ListAndByID` | List teams and fetch by ID |
| `TestStreams_ListAndByID` | List streams and fetch by ID |
| `TestStreamsWS_UpgradeAndTelemetryBroadcast` | WebSocket upgrade and telemetry broadcast |

#### Event Mutations (`events_mutations_test.go`)
| Function | Description |
|---|---|
| `TestCreateEventEndpoint` | Admin POST creates event |
| `TestUpdateEventEndpoint` | Admin PUT updates event fields |
| `TestDeleteEventEndpoint` | Admin DELETE removes event |

#### Admin Management (`admin_management_test.go`)
| Function | Description |
|---|---|
| `TestCategoryManagementAdminRoutes` | CRUD for categories via admin routes |
| `TestJourneyCatalogManagementAdminRoutes` | CRUD for journey catalog via admin routes |
| `TestTeamManagementAdminRoutes` | CRUD for teams via admin routes |
| `TestStreamManagementAdminRoutes` | CRUD for streams via admin routes |

#### Garage & Passes (`garage_passes_test.go`)
| Function | Description |
|---|---|
| `TestGetGarageYearsReturnsActualRange` | Returns min/max year range for make+model |
| `TestGetGarageVehicleReturnsSpec` | Returns computed vehicle spec |
| `TestPostGarageTuneAppliesConfig` | Applies tuning multiplier to base spec |
| `TestPurchasePassEndpoint` | Purchase reduces spotsLeft on valid pass |

#### Journeys & Teams (`journeys_teams_test.go`)
| Function | Description |
|---|---|
| `TestBookJourneyEndpoint` | Book journey stores booking for user |
| `TestFollowTeamEndpoint` | Follow team creates relation |
| `TestUnfollowTeamEndpoint` | Unfollow team removes relation |

#### Orders & Reminders (`orders_reminders_test.go`)
| Function | Description |
|---|---|
| `TestCreateOrderEndpoint` | POST creates pending order |
| `TestGetMyOrdersEndpoint` | GET returns authenticated user's orders |
| `TestGetOrderByIDEndpoint` | GET by ID returns correct order |
| `TestCancelOrderEndpoint` | DELETE cancels pending order |
| `TestSetReminderEndpoint` | POST creates reminder with future time |
| `TestGetMyRemindersEndpoint` | GET returns user reminders |
| `TestDeleteReminderEndpoint` | DELETE removes reminder |

#### Sprint 4 New Handlers (`additional_handlers_test.go`)
| Function | Description |
|---|---|
| `TestUpdateCurrentUserPlanAndRoleFlow` | `PUT /auth/me/plan` and `/auth/me/role` with plan-hierarchy validation |
| `TestCartHandlersFlow` | GET/POST/DELETE cart; dedup and clear |
| `TestPaymentHandlersFlow` | Create intent → confirm → get status |
| `TestNotificationHandlersFallback` | GET notifications and mark-read with in-memory fallback |
| `TestGarageConfigHandlersFlow` | Save, list, and delete garage configs |
| `TestTeamManagerHandlers` | Add and remove team managers |
| `TestTeamRelationsMemberAndSponsorHandlers` | Add/list/remove team members and sponsors |
| `TestPassesCatalogAndMyPassesAndJourneyBookings` | Pass catalog, my-passes, and my journey bookings |
| `TestGarageSupplementalHandlers` | Makes, models, trims, tuning-configs, and search endpoints |
| `TestStreamYoutubeAndAdminHandlers` | Streams list, YouTube search proxy, admin stream CRUD |

Run:

```bash
cd nitrous-backend
go test ./...
```

---

## Updated Backend API Documentation

### Base URL
- Local: `http://localhost:8080`
- API Prefix: `/api`

### Authentication
All protected endpoints require `Authorization: Bearer <token>` header. Token is returned from `/api/auth/login`.

---

### Health

#### `GET /health`
- Auth: Public
- Response: `{ "status": "ok", "message": "Nitrous API is running" }`

---

### Auth

#### `POST /api/auth/register`
- Auth: Public
- Body: `{ "email": "user@example.com", "password": "securepass", "name": "User" }`
- Notes: Always creates account with `role=viewer` and `plan=FREE`

#### `POST /api/auth/login`
- Auth: Public
- Body: `{ "email": "user@example.com", "password": "securepass" }`
- Response: `{ "token": "<jwt>", "user": { ... } }`

#### `GET /api/auth/me`
- Auth: Bearer token required
- Response: current user object (id, email, name, role, plan)

#### `PUT /api/auth/me/plan`
- Auth: Bearer token required
- Body: `{ "plan": "VIP" }` — valid values: `FREE`, `VIP`, `PLATINUM`
- Notes: Plan can only be upgraded (FREE < VIP < PLATINUM)

#### `PUT /api/auth/me/role`
- Auth: Bearer token required
- Body: `{ "role": "participant" }` — valid values: `viewer`, `participant`, `manager`, `sponsor`, `admin`
- Notes: Role change is validated against current plan (e.g., sponsor requires PLATINUM)

---

### Events

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/events` | Public | List all events; filter by `?category=slug` |
| GET | `/api/events/live` | Public | List live-only events |
| GET | `/api/events/:id` | Public | Get single event |
| POST | `/api/events` | Admin | Create event |
| PUT | `/api/events/:id` | Admin | Update event |
| DELETE | `/api/events/:id` | Admin | Delete event |

---

### Categories

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/categories` | Public | List all categories |
| GET | `/api/categories/:slug` | Public | Get category by slug |
| POST | `/api/categories` | Admin | Create category |
| PUT | `/api/categories/:slug` | Admin | Update category |
| DELETE | `/api/categories/:slug` | Admin | Delete category |

---

### Journeys

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/journeys` | Public | List all journeys |
| GET | `/api/journeys/:id` | Public | Get single journey |
| GET | `/api/journeys/my` | Bearer | List authenticated user's bookings |
| POST | `/api/journeys/:id/book` | Bearer | Book journey; body: `{ "quantity": 1 }` |
| POST | `/api/journeys` | Admin | Create journey |
| PUT | `/api/journeys/:id` | Admin | Update journey |
| DELETE | `/api/journeys/:id` | Admin | Delete journey |

---

### Access Passes

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/passes/catalog` | Public | List all passes with spotsLeft |
| GET | `/api/passes/my` | Bearer | List passes purchased by current user |
| POST | `/api/passes/:id/purchase` | Bearer | Purchase pass; body: `{ "quantity": 1 }` |

---

### Merch

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/merch` | Public | List all merch items |
| GET | `/api/merch/:id` | Public | Get single merch item |

---

### Orders

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/orders` | Bearer | List authenticated user's orders |
| POST | `/api/orders` | Bearer | Create order; body: `{ "itemId": "...", "quantity": 1 }` |
| GET | `/api/orders/:id` | Bearer | Get single order |
| DELETE | `/api/orders/:id` | Bearer | Cancel order (pending status only) |

---

### Payments

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/payments/intent` | Bearer | Create payment intent; body: `{ "amount": 4999, "currency": "usd", "description": "..." }` |
| POST | `/api/payments/confirm` | Bearer | Confirm payment; body: `{ "paymentId": "...", "paymentMethodId": "pm_..." }` |
| GET | `/api/payments/:id` | Bearer | Get payment status |
| GET | `/api/payments/my` | Bearer | List all payments for current user |

---

### Cart

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/cart` | Bearer | Get current user's cart items |
| POST | `/api/cart` | Bearer | Save cart; body: `{ "items": [{ "itemId": "...", "quantity": 1, "type": "merch" }] }` |
| DELETE | `/api/cart` | Bearer | Clear all cart items |

---

### Reminders

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/reminders` | Bearer | List reminders for current user |
| POST | `/api/reminders` | Bearer | Create reminder; body: `{ "eventId": "...", "remindAt": "2026-05-01T18:00:00Z" }` |
| DELETE | `/api/reminders/:id` | Bearer | Delete reminder |

---

### Notifications

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/notifications` | Bearer | List notifications for current user |
| POST | `/api/notifications/:id/read` | Bearer | Mark notification as read |

---

### Teams

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/teams` | Public | List all teams |
| GET | `/api/teams/:id` | Public | Get single team |
| POST | `/api/teams/:id/follow` | Bearer | Follow team |
| POST | `/api/teams/:id/unfollow` | Bearer | Unfollow team |
| POST | `/api/teams` | Admin | Create team |
| PUT | `/api/teams/:id` | Admin | Update team |
| DELETE | `/api/teams/:id` | Admin | Delete team |
| GET | `/api/teams/:id/members` | Bearer | List team members |
| POST | `/api/teams/:id/members` | Manager/Admin | Add team member |
| DELETE | `/api/teams/:id/members/:userId` | Manager/Admin | Remove team member |
| GET | `/api/teams/:id/sponsors` | Bearer | List team sponsors |
| POST | `/api/teams/:id/sponsors` | Manager/Admin | Add sponsor |
| DELETE | `/api/teams/:id/sponsors/:userId` | Manager/Admin | Remove sponsor |
| GET | `/api/teams/:id/managers` | Bearer | List team managers |
| POST | `/api/teams/:id/managers` | Admin | Add team manager |
| DELETE | `/api/teams/:id/managers/:userId` | Admin | Remove team manager |

---

### Streams

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/streams` | Public | List all streams |
| GET | `/api/streams/:id` | Public | Get single stream |
| GET | `/api/streams/ws` | Public | WebSocket connection for real-time telemetry |
| GET | `/api/streams/openf1/sessions` | Public | List OpenF1 sessions |
| GET | `/api/streams/openf1/sessions/:sessionKey/telemetry` | Public | Telemetry for session |
| POST | `/api/streams` | Admin | Create stream |
| PUT | `/api/streams/:id` | Admin | Update stream |
| DELETE | `/api/streams/:id` | Admin | Delete stream |

---

### Garage

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/garage/makes` | Public | List vehicle makes (NHTSA) |
| GET | `/api/garage/models?make=TOYOTA` | Public | List models for make |
| GET | `/api/garage/years?make=TOYOTA&model=Camry` | Public | Year range for make+model |
| GET | `/api/garage/trims?make=TOYOTA&model=Camry&year=2024` | Public | Trim list |
| GET | `/api/garage/vehicle?make=TOYOTA&model=Camry&year=2024` | Public | Computed vehicle spec |
| GET | `/api/garage/tuning-configs` | Public | Available tuning profiles |
| POST | `/api/garage/tune` | Public | Apply tuning; body: `{ "make": "TOYOTA", "model": "Camry", "year": 2024, "tuning": "track" }` |
| GET | `/api/garage/search?q=toyota` | Public | Search makes by term |
| GET | `/api/garage/configs` | Bearer | List saved garage configs for current user |
| POST | `/api/garage/configs` | Bearer | Save a garage config; body: `{ "name": "My Setup", "config": { ... } }` |
| DELETE | `/api/garage/configs/:id` | Bearer | Delete saved garage config |

---

### YouTube (Live page proxy)

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/youtube/search?q=f1+race` | Public | Proxy YouTube Data API search; returns video list |

---

## Verification Commands

```bash
# Backend tests
cd nitrous-backend
go test ./...

# Frontend unit tests
cd nitrous-app
npm run test

# Frontend Cypress E2E (app must be running)
cd nitrous-app
npx cypress run

# Frontend Cypress component tests
cd nitrous-app
npx cypress run --component
```
