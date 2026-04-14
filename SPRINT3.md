# Sprint 3

## Sprint 3 Work Completed

### 1. Garage Feature Upgrade (NHTSA Integration)
- Migrated Garage vehicle discovery flow to NHTSA vPIC-backed endpoints.
- Implemented backend garage handlers for:
	- makes lookup
	- models lookup (by make, optional year)
	- year range detection
	- trims lookup
	- vehicle spec generation
	- tuning application
	- garage search
- Updated year-range behavior to return the actual available min/max years for a selected make/model to prevent invalid-year blank data in the frontend.
- Preserved NHTSA-only data source behavior in garage API flow.

### 2. Garage Frontend UX and Data Flow
- Reworked Garage page to fetch make/model/year options from backend API routes.
- Added robust fetch error handling and status messaging for vehicle spec requests.
- Added saved garage configurations (make/model/year + engine snapshot) with quick reselect behavior.
- Removed unnecessary engine selector UI when not actionable and kept engine display in HUD/spec context.
- Repaired merge-damage in garage frontend module and restored compile-safe state/effect logic.

### 3. Backend Stability and Handler Fixes
- Fixed `passes.go` to match current backend architecture:
	- removed stale SQL/transaction usage (`database.GetDB`)
	- aligned auth identity access to middleware context (`userID`)
	- migrated pass purchase operation to in-memory data model with locking
- Added in-memory pass catalog + purchase tracking seed structures in database layer.
- Removed stale garage route registration reference in `main.go` that no longer existed after route inlining.
- Verified backend compile integrity with `go test ./...`.

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

## Garage API (Sprint 3 Focus)

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
{ "minYear": "1980", "maxYear": "2024", "source": "nhtsa" }
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

