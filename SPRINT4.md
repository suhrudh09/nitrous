# Sprint 4

## Sprint 4 Work Completed

### 1. Backend — PostgreSQL migration & DB work
- Migrated backend database flow to PostgreSQL and moved seed/load flow to a postgres-first provider persistence model.
- Files added/updated: `nitrous-backend/database/db.go`, `nitrous-backend/database/schema.sql` (added), `nitrous-backend/database/seed.sql` (added), updated `nitrous-backend/go.mod` and `nitrous-backend/go.sum`.
- Updated handlers to use the new DB layer: `nitrous-backend/handlers/auth.go`, `events.go`, `orders.go`, `other.go`, `passes.go`, `reminders.go`, `teams.go`, and new/updated handler files for provider/team management.
- Notes: schema and seed files are provided in `nitrous-backend/database/` and should be applied to a local Postgres instance before running the backend.

### 2. Backend — Role-based access, team management, and integration
- Enforced role-based access controls and added team relationship management handlers.
- Files added/modified: `nitrous-backend/handlers/admin.go` (new), `team_managers.go` (new), `team_relations.go` (new), and updates to `handlers/teams.go`, `middleware/auth.go`, and `models/models.go`.

### 3. Backend — CORS and YouTube embed logic
- Updated CORS to allow the Vercel frontend domain.
- Files modified: `nitrous-backend/main.go`.
- Added backend support for YouTube embed/video search on the live page: `nitrous-backend/handlers/youtube.go` and `nitrous-backend/config/config.go` updates.

### 4. Frontend — Payment, Cart, Orders, and UI work
- Integrated payment gateway and added payment page/components.
- Files added/modified: `nitrous-app/app/payment/page.tsx` and `payment.module.css` (new), updates to `nitrous-app/app/cart/page.tsx`, `nitrous-app/app/merch/page.tsx`, `nitrous-app/app/orders/page.tsx` and related styles/components.

### 5. Frontend — Tooling and tests
- Upgraded ESLint and resolved a `next-config` dependency conflict (`nitrous-app/package.json`, `nitrous-app/package-lock.json`).
- Added `cypress.config.ts` to TypeScript exclude list to avoid type-checking test config (`nitrous-app/tsconfig.json`).
- Added/updated Cypress support and component test artifacts in `nitrous-app/cypress/` and component tests for `Hero` and `Nav`.

### 6. UX / Misc frontend fixes
- Garage config save/load behavior fixed and signup role selection default fixed to `viewer` when intended.
- Multiple UI fixes to live/garage/hero/navigation components and CSS modules.

---

## Frontend Unit Tests

Current frontend unit test suite (Jest):

- `nitrous-app/__tests__/api.test.ts`

### Frontend Unit Test Coverage in `api.test.ts`
- `getEvents`
	- fetch success path
	- API failure handling
- `getCategories`
	- fetch success path
- `getJourneys`
	- fetch success path
- `getMerchItems`
	- fetch success path
- `getEventById`
	- fetch by ID success path
	- endpoint path assertion
- Authentication API functions
	- `register` success path
	- `login` success path
	- `getCurrentUser` success path with bearer token
- Error handling behavior
	- API error propagation
	- network error propagation

Run command:

```bash
cd nitrous-app
npm run test
```

---

## Backend Unit Tests

Current backend test files:

- `nitrous-backend/handlers/auth_handlers_test.go`
- `nitrous-backend/handlers/admin_management_test.go`
- `nitrous-backend/handlers/events_mutations_test.go`
- `nitrous-backend/handlers/garage_passes_test.go`
- `nitrous-backend/handlers/handlers_test.go`
- `nitrous-backend/handlers/journeys_teams_test.go`
- `nitrous-backend/handlers/orders_reminders_test.go`
- `nitrous-backend/handlers/test_helpers_test.go`
- `nitrous-backend/middleware/auth_test.go`
- `nitrous-backend/middleware/admin_test.go`
- `nitrous-backend/utils/jwt_test.go`

### Backend Test Functions (Highlights)

#### Auth & JWT
- `TestRegisterFlow`
- `TestLoginFlow`
- `TestGetCurrentUserFlow`
- `TestJWTUtility`

#### Middleware
- `TestAuthMiddleware`
- `TestAdminMiddleware`

#### Events/Categories/Journeys/Teams/Streams (read + mutation)
- `TestGetEvents_ListAndCategoryFilter`
- `TestGetLiveEvents_ReturnsOnlyLive`
- `TestGetEventByID_FoundAndNotFound`
- `TestCategories_ListAndBySlug`
- `TestJourneys_ListAndByID`
- `TestMerch_ListAndByID`
- `TestTeams_ListAndByID`
- `TestStreams_ListAndByID`
- `TestStreamsWS_UpgradeAndTelemetryBroadcast`
- `TestCreateEventEndpoint`
- `TestUpdateEventEndpoint`
- `TestDeleteEventEndpoint`

#### Admin Management Routes
- `TestCategoryManagementAdminRoutes`
- `TestJourneyCatalogManagementAdminRoutes`
- `TestTeamManagementAdminRoutes`
- `TestStreamManagementAdminRoutes`

#### User Action Flows
- `TestBookJourneyEndpoint`
- `TestFollowTeamEndpoint`
- `TestUnfollowTeamEndpoint`
- `TestCreateOrderEndpoint`
- `TestGetMyOrdersEndpoint`
- `TestGetOrderByIDEndpoint`
- `TestCancelOrderEndpoint`
- `TestSetReminderEndpoint`
- `TestGetMyRemindersEndpoint`
- `TestDeleteReminderEndpoint`

#### Sprint 4 New/Updated Functionality Tests
- Tests updated to reflect Postgres migration and role/team changes (garage/tune/passes tests adjusted to remove InitDB fixture coupling).

Run command:

```bash
cd nitrous-backend
go test ./...
```

---

## Updated Backend API Documentation

### Base URL
- Local: `http://localhost:8080`
- API Prefix: `/api`

### Health

#### GET `/health`
- Auth: Public
- Response:

```json
{ "status": "ok", "message": "Nitrous API is running" }
```

---

## Garage API (Sprint 4 focus — behavior unchanged)

### GET `/api/garage/makes`
- Auth: Public
- Purpose: List available vehicle makes from NHTSA.
- Response:

```json
{
	"makes": [
		{ "id": "448", "name": "TOYOTA", "country": "N/A" }
	]
}
```

### GET `/api/garage/models?make={make}[&year={year}]`
- Auth: Public
- Purpose: List models for make (optional year-filtered).
- Response:

```json
{
	"models": [
		{ "id": "2469", "name": "Camry", "make": "TOYOTA" }
	]
}
```

### GET `/api/garage/years?make={make}&model={model}`
- Auth: Public
- Purpose: Return actual available year range for selected make/model.
- Response:

```json
{ "minYear": "1980", "maxYear": "2026", "source": "nhtsa" }
```

### GET `/api/garage/trims?make={make}&model={model}&year={year}`
- Auth: Public
- Purpose: Return trims for selected vehicle.
- Response:

```json
{
	"trims": [{ "model_trim": "Camry", "model_year": "2024" }],
	"source": "nhtsa"
}
```

### GET `/api/garage/vehicle?make={make}&model={model}&year={year}`
- Auth: Public
- Purpose: Return computed vehicle spec used by garage HUD/cards.
- Response:

```json
{
	"vehicle": {
		"make": "TOYOTA",
		"model": "Camry",
		"year": 2024,
		"hp": 280,
		"torque": 300,
		"topSpeed": 155,
		"weight": 3400,
		"zeroToSixty": 6.4,
		"engine": "N/A"
	},
	"source": "nhtsa"
}
```

### GET `/api/garage/tuning-configs`
- Auth: Public
- Purpose: Return available tuning profiles and multipliers.

### POST `/api/garage/tune`
- Auth: Public
- Purpose: Apply selected tuning profile to a base vehicle spec.
- Body:

```json
{ "make": "TOYOTA", "model": "Camry", "year": 2024, "tuning": "track" }
```

- Response:

```json
{
	"base": { "hp": 280, "topSpeed": 155 },
	"tuned": { "hp": 330, "topSpeed": 170, "config": "Track" },
	"delta": { "hp": 50, "topSpeed": 15 },
	"config": { "label": "Track", "hpMult": 1.18 }
}
```

### GET `/api/garage/search?q={term}`
- Auth: Public
- Purpose: Search makes by string term.

---

## Auth

### POST `/api/auth/register`
- Auth: Public
- Body:

```json
{ "email": "user@example.com", "password": "securepass123", "name": "User" }
```

### POST `/api/auth/login`
- Auth: Public
- Body:

```json
{ "email": "user@example.com", "password": "securepass123" }
```

### GET `/api/auth/me`
- Auth: Bearer token required

---

## Events

- `GET /api/events`
- `GET /api/events/live`
- `GET /api/events/:id`
- `POST /api/events` (Admin)
- `PUT /api/events/:id` (Admin)
- `DELETE /api/events/:id` (Admin)

## Categories

- `GET /api/categories`
- `GET /api/categories/:slug`
- `POST /api/categories` (Admin)
- `PUT /api/categories/:slug` (Admin)
- `DELETE /api/categories/:slug` (Admin)

## Journeys

- `GET /api/journeys`
- `GET /api/journeys/:id`
- `POST /api/journeys` (Admin)
- `PUT /api/journeys/:id` (Admin)
- `DELETE /api/journeys/:id` (Admin)
- `POST /api/journeys/:id/book` (Authenticated)

## Merch

- `GET /api/merch`
- `GET /api/merch/:id`

## Teams

- `GET /api/teams`
- `GET /api/teams/:id`
- `POST /api/teams` (Admin)
- `PUT /api/teams/:id` (Admin)
- `DELETE /api/teams/:id` (Admin)
- `POST /api/teams/:id/follow` (Authenticated)
- `POST /api/teams/:id/unfollow` (Authenticated)

## Streams

- `GET /api/streams`
- `GET /api/streams/openf1/sessions`
- `GET /api/streams/openf1/sessions/:sessionKey/telemetry`
- `GET /api/streams/:id`
- `POST /api/streams` (Admin)
- `PUT /api/streams/:id` (Admin)
- `DELETE /api/streams/:id` (Admin)
- `GET /api/streams/ws`

## Reminders

- `GET /api/reminders` (Authenticated)
- `POST /api/reminders` (Authenticated)
- `DELETE /api/reminders/:id` (Authenticated)

## Orders

- `GET /api/orders` (Authenticated)
- `POST /api/orders` (Authenticated)
- `GET /api/orders/:id` (Authenticated)
- `DELETE /api/orders/:id` (Authenticated)

## Passes

### POST `/api/passes/:id/purchase`
- Auth: Bearer token required
- Purpose: Reserve and purchase event pass if spots remain.
- Success:

```json
{ "message": "Pass purchased successfully", "passId": "pass-daytona-grandstand" }
```

---

## Verification Commands

```bash
# Backend tests
cd nitrous-backend
go test ./...

# Frontend unit tests
cd ../nitrous-app
npm run test
```
