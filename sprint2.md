# Sprint 2

## Sprint 2 Work Completed

## Frontend Testing

### Frontend Unit Tests
Unit tests created using Jest and React Testing Library with comprehensive coverage of components and API utilities.

#### Unit Test Files Added
- `nitrous-app/__tests__/Nav.test.tsx` — Navigation component tests
- `nitrous-app/__tests__/Hero.test.tsx` — Hero section component tests
- `nitrous-app/__tests__/api.test.ts` — API utility function tests

#### Frontend Unit Tests Implemented
**Navigation Component (Nav.test.tsx)**
- Renders navigation with logo
- Renders all navigation links (Live, Events, Teams, Journeys, Merch)
- Renders Sign In button
- Displays live events count ("4 Events Live")
- Verifies navigation structure and href attributes

**Hero Component (Hero.test.tsx)**
- Renders hero section
- Displays main title with "NITROUS" and "FUEL"
- Displays subtitle text
- Renders action buttons (Ignite Stream, Explore Events)
- Renders hero navigation cards
- Renders background image
- Contains circuit layer elements
- Renders HUD corners with SVG elements
- Displays HUD label text

**API Utility Functions (api.test.ts)**
- **Event API tests:**
  - `getEvents()` - fetches and returns events array
  - `getEventById(id)` - fetches event by ID
  - Handles errors properly
- **Category API tests:**
  - `getCategories()` - fetches all categories
- **Journey API tests:**
  - `getJourneys()` - fetches all journeys
- **Merch API tests:**
  - `getMerchItems()` - fetches merch items
- **Authentication tests:**
  - `register()` - user registration with email, password, name
  - `login()` - user login with email and password
  - `getCurrentUser()` - retrieves current user with JWT token
- **Error handling:**
  - API error handling for failed requests
  - Network error handling

### Frontend E2E Tests (Cypress)
End-to-end tests created using Cypress framework with real user interaction scenarios.

#### Cypress Test Files Added
- `nitrous-app/cypress/cypress.config.ts` — Cypress configuration
- `nitrous-app/cypress/e2e/home.cy.ts` — Home page navigation tests
- `nitrous-app/cypress/e2e/hero-interactions.cy.ts` — Hero section interaction tests

#### Frontend E2E Tests Implemented
**Home Page Navigation (home.cy.ts)**
- Loads home page successfully
- Displays navigation menu with all links (Live, Events, Teams, Journeys, Merch)
- Displays Sign In button
- Displays hero title and subtitle
- Displays action buttons
- Navigation: Events page
- Navigation: Teams page
- Navigation: Journeys page
- Navigation: Merch page
- Displays live events status

**Hero Section Interactions (hero-interactions.cy.ts)**
- Verifies hero action buttons are clickable
- Verifies all navigation cards are visible (Garage, Event Passes, Live Streams, Teams, Journeys, Merch)
- Click Ignite Stream button
- Click Explore Events button
- Navigate through hero nav cards
- Navigate to live streams from hero
- Verifies hero section styling elements

### Test Coverage Summary
- **Component Tests:** 2 components (Nav, Hero)
- **API Tests:** 9 functions tested (Events, Categories, Journeys, Merch, Auth)
- **E2E Tests:** 2 test suites with 15+ test cases

### Running Tests

#### Unit Tests
```bash
# Run all unit tests
npm run test

# Watch mode
npm run test:watch

# Coverage report
npm run test:coverage
```

#### Cypress E2E Tests
```bash
# Open Cypress Test Runner
npm run cypress

# Run Cypress tests headless
npm run cypress:run
```


### 1. Backend Integration and Authorization Improvements
- Implemented role-based authorization (`admin` vs `user`) for management operations.
- Added `Role` field to user model.
- Added admin authorization middleware.
- Applied admin-only protection to management routes for:
  - Event management
  - Category management
  - Journey catalog management
  - Team content management
  - Stream content management
- Seeded a default admin user in in-memory database seed data.

### 2. Backend Feature Work Added
- Added category management handlers:
  - Create category
  - Update category
  - Delete category
- Added journey catalog management handlers:
  - Create journey
  - Update journey
  - Delete journey
- Added team content management handlers:
  - Create team
  - Update team
  - Delete team
- Added stream content management handlers:
  - Create stream
  - Update stream
  - Delete stream

### 3. Backend Unit Tests Added
- Added and validated backend tests using Go's testing framework and Gin test router patterns.
- Test command used:
  - `go test ./...`

#### Test files added
- `nitrous-backend/handlers/auth_handlers_test.go`
- `nitrous-backend/handlers/admin_management_test.go`
- `nitrous-backend/handlers/handlers_test.go`
- `nitrous-backend/handlers/events_mutations_test.go`
- `nitrous-backend/handlers/journeys_teams_test.go`
- `nitrous-backend/handlers/orders_reminders_test.go`
- `nitrous-backend/handlers/test_helpers_test.go`
- `nitrous-backend/middleware/auth_test.go`
- `nitrous-backend/middleware/admin_test.go`
- `nitrous-backend/utils/jwt_test.go`

#### Backend tests implemented
- Authentication
  - `TestRegisterFlow`
  - `TestLoginFlow`
  - `TestGetCurrentUserFlow`
- Middleware
  - `TestAuthMiddleware`
  - `TestAdminMiddleware`
- Event mutation routes
  - `TestCreateEventEndpoint`
  - `TestUpdateEventEndpoint`
  - `TestDeleteEventEndpoint`
- Public/read route coverage
  - `TestGetEvents_ListAndCategoryFilter`
  - `TestGetLiveEvents_ReturnsOnlyLive`
  - `TestGetEventByID_FoundAndNotFound`
  - `TestCategories_ListAndBySlug`
  - `TestJourneys_ListAndByID`
  - `TestMerch_ListAndByID`
  - `TestTeams_ListAndByID`
  - `TestStreams_ListAndByID`
  - `TestStreamsWS_UpgradeAndTelemetryBroadcast`
- Admin management routes
  - `TestCategoryManagementAdminRoutes`
  - `TestJourneyCatalogManagementAdminRoutes`
  - `TestTeamManagementAdminRoutes`
  - `TestStreamManagementAdminRoutes`
- Journey and team routes
  - `TestBookJourneyEndpoint`
  - `TestFollowTeamEndpoint`
  - `TestUnfollowTeamEndpoint`
- Orders and reminders
  - `TestCreateOrderEndpoint`
  - `TestGetMyOrdersEndpoint`
  - `TestGetOrderByIDEndpoint`
  - `TestSetReminderEndpoint`
  - `TestGetMyRemindersEndpoint`
  - `TestDeleteReminderEndpoint`
- JWT utility
  - `TestJWTUtility`

## Backend API Documentation

### Base URL
- Local: `http://localhost:8080`
- API prefix: `/api`

### Authentication Model
- JWT Bearer token is required for protected routes.
- Header format:
  - `Authorization: Bearer <token>`
- Role-based authorization:
  - Admin-only routes require a valid user with role `admin`.

---

## Health

### GET `/health`
- Auth: Public
- Purpose: API health check
- Response:
```json
{ "status": "ok", "message": "Nitrous API is running" }
```

---

## Auth

### POST `/api/auth/register`
- Auth: Public
- Body:
```json
{ "email": "user@example.com", "password": "securepass123", "name": "User Name" }
```
- Behavior:
  - Creates a new user with default role `user`
  - Returns JWT token
- Success: `201 Created`
- Errors: `400 Bad Request`, `409 Conflict`

### POST `/api/auth/login`
- Auth: Public
- Body:
```json
{ "email": "user@example.com", "password": "securepass123" }
```
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`

### GET `/api/auth/me`
- Auth: Protected
- Success: `200 OK` (current user object)
- Errors: `401 Unauthorized`, `404 Not Found`

---

## Events

### GET `/api/events`
- Auth: Public
- Query: optional `category`
- Success: `200 OK`

### GET `/api/events/live`
- Auth: Public
- Success: `200 OK`

### GET `/api/events/:id`
- Auth: Public
- Success: `200 OK`
- Error: `404 Not Found`

### POST `/api/events`
- Auth: Protected + Admin
- Purpose: Create event
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`

### PUT `/api/events/:id`
- Auth: Protected + Admin
- Purpose: Update event
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### DELETE `/api/events/:id`
- Auth: Protected + Admin
- Purpose: Delete event
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

---

## Categories

### GET `/api/categories`
- Auth: Public
- Success: `200 OK`

### GET `/api/categories/:slug`
- Auth: Public
- Success: `200 OK`
- Error: `404 Not Found`

### POST `/api/categories`
- Auth: Protected + Admin
- Purpose: Create category
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`

### PUT `/api/categories/:slug`
- Auth: Protected + Admin
- Purpose: Update category
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### DELETE `/api/categories/:slug`
- Auth: Protected + Admin
- Purpose: Delete category
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

---

## Journeys

### GET `/api/journeys`
- Auth: Public
- Success: `200 OK`

### GET `/api/journeys/:id`
- Auth: Public
- Success: `200 OK`
- Error: `404 Not Found`

### POST `/api/journeys`
- Auth: Protected + Admin
- Purpose: Create journey (catalog management)
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`

### PUT `/api/journeys/:id`
- Auth: Protected + Admin
- Purpose: Update journey (catalog management)
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### DELETE `/api/journeys/:id`
- Auth: Protected + Admin
- Purpose: Delete journey (catalog management)
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### POST `/api/journeys/:id/book`
- Auth: Protected
- Purpose: Book journey (user action)
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`

---

## Teams

### GET `/api/teams`
- Auth: Public
- Success: `200 OK`

### GET `/api/teams/:id`
- Auth: Public
- Success: `200 OK`
- Error: `404 Not Found`

### POST `/api/teams`
- Auth: Protected + Admin
- Purpose: Create team
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`

### PUT `/api/teams/:id`
- Auth: Protected + Admin
- Purpose: Update team
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### DELETE `/api/teams/:id`
- Auth: Protected + Admin
- Purpose: Delete team
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### POST `/api/teams/:id/follow`
- Auth: Protected
- Purpose: Follow team (user action)
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`

### POST `/api/teams/:id/unfollow`
- Auth: Protected
- Purpose: Unfollow team (user action)
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`

---

## Streams

### GET `/api/streams`
- Auth: Public
- Success: `200 OK`

### GET `/api/streams/:id`
- Auth: Public
- Success: `200 OK`
- Error: `404 Not Found`

### POST `/api/streams`
- Auth: Protected + Admin
- Purpose: Create stream
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`

### PUT `/api/streams/:id`
- Auth: Protected + Admin
- Purpose: Update stream
- Success: `200 OK`
- Errors: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### DELETE `/api/streams/:id`
- Auth: Protected + Admin
- Purpose: Delete stream
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

### GET `/api/streams/ws`
- Auth: Public
- Purpose: WebSocket telemetry stream

---

## Merch

### GET `/api/merch`
- Auth: Public
- Success: `200 OK`

### GET `/api/merch/:id`
- Auth: Public
- Success: `200 OK`
- Error: `404 Not Found`

---

## Orders

### POST `/api/orders`
- Auth: Protected
- Purpose: Create order
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`

### GET `/api/orders`
- Auth: Protected
- Purpose: List current user's orders
- Success: `200 OK`
- Errors: `401 Unauthorized`

### GET `/api/orders/:id`
- Auth: Protected
- Purpose: Get current user's order by ID
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

---

## Reminders

### POST `/api/reminders`
- Auth: Protected
- Purpose: Create reminder
- Success: `201 Created`
- Errors: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`

### GET `/api/reminders`
- Auth: Protected
- Purpose: List current user's reminders
- Success: `200 OK`
- Errors: `401 Unauthorized`

### DELETE `/api/reminders/:id`
- Auth: Protected
- Purpose: Delete current user's reminder by ID
- Success: `200 OK`
- Errors: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`

---
