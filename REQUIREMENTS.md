## Software Requirements Specification (SRS)  
### OSINT Research Web Application (Inspired by Bellingcat)

---

### 1. Introduction

#### 1.1 Purpose  
This document specifies the functional and non‑functional requirements for a web‑based Open Source Intelligence (OSINT) research platform. The application will enable analysts, journalists, and researchers to collect, organise, analyse, visualise, and collaborate on publicly available information in a structured and secure manner, following methodologies popularised by organisations such as Bellingcat.

#### 1.2 Scope  
The system provides a multi‑user environment where investigators can create **cases**, ingest **evidence** (URLs, files, notes, raw data), enrich it with metadata, extract **entities**, and perform **geospatial** and **temporal** analysis. Collaboration features include commenting, team sharing, and activity logs. The platform is intended for lawful, open‑source intelligence gathering only.

#### 1.3 Definitions  
- **OSINT**: Open Source Intelligence – information collected from publicly available sources.  
- **Case**: A logical container representing an investigation or project.  
- **Evidence**: A single piece of information (text, URL, image, video, document, raw data) associated with a case.  
- **Entity**: A named object (person, organisation, location) extracted from evidence.  
- **Analyst**: A user with full create/edit/delete permissions within cases they are a member of.  
- **Viewer**: A user with read‑only access to assigned cases.

---

### 2. User Roles & Permissions

| Capability                         | Admin | Analyst | Viewer |
|------------------------------------|-------|---------|--------|
| Create/manage users                | ✅    | ❌      | ❌     |
| Create cases                       | ✅    | ✅      | ❌     |
| Edit/delete own cases              | ✅    | ✅      | ❌     |
| Add/edit/delete evidence in a case | ✅    | ✅*     | ❌     |
| Comment on cases/evidence          | ✅    | ✅      | ✅**   |
| View public or shared cases        | ✅    | ✅      | ✅     |
| Export reports                     | ✅    | ✅      | ✅***  |
| Manage case members                | ✅    | ✅****  | ❌     |

\* Analyst must be a member of the case with `editor` or `owner` role.  
\** Viewer can comment only if allowed by case owner.  
\*** Viewer may export if case is public or they have been granted export permission.  
\**** Analyst can manage members only if they are the case owner.

---

### 3. Functional Requirements

#### 3.1 Authentication & User Management
- [x] FR‑AUTH‑1: Users must register with a valid email and password.  
- [ ] FR‑AUTH‑2: Password must be at least 8 characters and include letters and numbers.  
- [x] FR‑AUTH‑3: Passwords must be hashed using bcrypt before storage.  
- [x] FR‑AUTH‑4: Login returns a short‑lived JWT access token (15 min) and a refresh token (7 days).  
- [x] FR‑AUTH‑5: Refresh token can be used to obtain a new access token without re‑authentication.  
- [ ] FR‑AUTH‑6: Users can reset their password via email link.  
- [ ] FR‑AUTH‑7: Admin can view, deactivate, or promote users.

#### 3.2 Case Management
- [x] FR‑CASE‑1: Authenticated users can create a new case with: title, description, status (`open`, `closed`, `archived`), tags, and optional location.  
- [x] FR‑CASE‑2: Case owner can edit case details, change status, add/remove tags, and delete the case.  
- [ ] FR‑CASE‑3: Cases can be marked as `private`, `shared` (specific users/groups), or `public`.  
- [x] FR‑CASE‑4: Owners can invite users by email or username and assign them a role within the case (`viewer`, `editor`, `owner`).  
- [x] FR‑CASE‑5: A case must have at least one owner at all times.  
- [ ] FR‑CASE‑6: Users can view a list of cases they own, are members of, or that are public, with filters by status, tags, and date.

#### 3.3 Evidence Management
- [x] FR‑EVD‑1: Within a case, users with edit rights can add evidence of the following types:  
  - Text note  
  - URL  
  - Image file  
  - Video file  
  - Document (PDF, DOCX, etc.)  
  - Raw data (JSON or other structured text)  
- [x] FR‑EVD‑2: For URL evidence, the system automatically fetches metadata (title, description, preview image) via server‑side scraping or oEmbed.  
- [x] FR‑EVD‑3: All evidence can have manually entered metadata: title, description, source, date (can be approximate), location (lat/lng), tags.  
- [x] FR‑EVD‑4: File uploads are stored on the server filesystem or configured cloud storage; the database stores the file path/URL and metadata.  
- [x] FR‑EVD‑5: Evidence can be edited or deleted by users with proper permissions. Deleting evidence also removes associated entities, comments, and file (if any).  
- [x] FR‑EVD‑6: Evidence can be filtered by type, date range, tags, location radius, and full‑text search.

#### 3.4 Search & Filtering
- [x] FR‑SRCH‑1: Global search across cases and evidence using PostgreSQL full‑text search or `ILIKE` on title, description, and text content.  
- [x] FR‑SRCH‑2: Search results are paginated (default 20 per page).  
- [x] FR‑SRCH‑3: Filters can be combined: type, date range, tags, location (point + radius), custom fields.  
- [ ] FR‑SRCH‑4: Advanced search allows searching for evidence that contains a specific extracted entity.

#### 3.5 Geospatial Analysis
- [x] FR‑GEO‑1: Evidence items may have geographic coordinates (WGS84).  
- [x] FR‑GEO‑2: Within a case, all geo‑tagged evidence is displayed on an interactive map (Leaflet + OpenStreetMap).  
- [ ] FR‑GEO‑3: Map markers cluster when zoomed out.  
- [ ] FR‑GEO‑4: Users can draw polygons or circles on the map to filter evidence located inside the shape.  
- [x] FR‑GEO‑5: An endpoint returns evidence as GeoJSON for client‑side rendering.  
- [x] FR‑GEO‑6: The database uses PostGIS to perform spatial queries (e.g., `ST_Within`, `ST_Distance`).

#### 3.6 Timeline Analysis
- [ ] FR‑TIME‑1: Evidence can have a `date` field (exact or approximate; approximate dates are stored with a precision flag).  
- [x] FR‑TIME‑2: The case detail page includes a timeline view listing all evidence chronologically.  
- [x] FR‑TIME‑3: Timeline can be filtered by date range and evidence type.  
- [ ] FR‑TIME‑4: Clicking a timeline item opens the full evidence detail.

#### 3.7 Entity Extraction
- [x] FR‑ENT‑1: The system provides an endpoint to run entity extraction on text‑based evidence (notes, URL descriptions, document text).  
- [x] FR‑ENT‑2: Basic extraction uses regex/NLP to identify persons, organisations, and locations.  
- [x] FR‑ENT‑3: Extracted entities are stored in a separate table and linked to the source evidence.  
- [ ] FR‑ENT‑4: Entities are displayed as clickable tags; clicking shows all evidence in the case containing that entity.  
- [ ] FR‑ENT‑5: Users can manually add or remove entity links.

#### 3.8 Collaboration & Comments
- [x] FR‑COL‑1: Users can post comments on cases and on individual evidence items.  
- [x] FR‑COL‑2: Comments support plain text and `@username` mentions.  
- [ ] FR‑COL‑3: When a user is mentioned, they receive an in‑app notification (and optionally an email).  
- [ ] FR‑COL‑4: Comments are displayed chronologically with author, timestamp, and edit/delete options for the author or case owner.  
- [ ] FR‑COL‑5: All actions (create, update, delete) are logged in an activity feed.

#### 3.9 Export & Reporting
- [x] FR‑EXP‑1: Users can export a case as a PDF report containing: case summary, list of evidence (with metadata), a map snapshot (if any geo‑data), and a timeline.  
- [x] FR‑EXP‑2: Users can export the filtered evidence list as CSV.  
- [x] FR‑EXP‑3: PDF generation is performed server‑side using a Go library (e.g., `gofpdf`, `unidoc`) or via headless browser rendering of HTML.  
- [ ] FR‑EXP‑4: Export buttons respect user permissions (viewer can export only if allowed).

#### 3.10 Dashboard
- [x] FR‑DASH‑1: After login, the user is redirected to a dashboard showing:  
  - Recent cases (last 10 updated)  
  - Total counts (cases, evidence, entities) for the user  
  - Activity feed (last 20 actions)  
  - Quick “Create Case” button  
- [x] FR‑DASH‑2: Dashboard data is fetched via a single aggregated API endpoint for performance.

---

### 4. Non‑Functional Requirements

#### 4.1 Performance
- [ ] 4.1.1 API response time for typical queries < 300 ms under normal load.  
- [ ] 4.1.2 The system should support at least 50 concurrent active users without degradation.  
- [ ] 4.1.3 Map rendering with up to 5,000 evidence points should remain responsive (using clustering).

#### 4.2 Security
- [ ] 4.2.1 All communication must use HTTPS (TLS 1.2+).  
- [x] 4.2.2 JWT tokens must be signed with a strong secret (HS256) or RSA.  
- [x] 4.2.3 All user inputs must be validated and sanitised to prevent XSS, SQL injection, and path traversal.  
- [ ] 4.2.4 File uploads must be restricted by MIME type and size (max 50 MB).  
- [x] 4.2.5 API must implement rate limiting (e.g., 100 requests/minute per user).  
- [x] 4.2.6 CORS must be configured to allow only the frontend origin.  
- [x] 4.2.7 Sensitive data (passwords, tokens) must never be logged.

#### 4.3 Usability
- [x] 4.3.1 The UI must be responsive and work on desktop and tablet screens.  
- [x] 4.3.2 All forms must provide inline validation and clear error messages.  
- [x] 4.3.3 The interface should use familiar design patterns (e.g., sidebar navigation, tabs) to reduce learning curve.

#### 4.4 Scalability
- [x] 4.4.1 The backend must be stateless to allow horizontal scaling (except file storage which can be offloaded).  
- [x] 4.4.2 Database queries must use indexes on frequently queried fields (case_id, tags, dates, geospatial columns).  
- [x] 4.4.3 The architecture should allow adding new evidence types or analysis modules with minimal changes.

#### 4.5 Reliability & Maintainability
- [x] 4.5.1 All backend errors must be logged with stack traces and request context.  
- [ ] 4.5.2 Database migrations must be versioned and reversible where possible.  
- [ ] 4.5.3 Code must follow standard Go and React best practices (linting, formatting).  
- [x] 4.5.4 The system should have automated tests for critical paths.

#### 4.6 Compliance (GDPR)
- [ ] 4.6.1 Users can request deletion of their account and all associated data.  
- [ ] 4.6.2 The platform must provide a way to export all personal data.  
- [ ] 4.6.3 Cookies / local storage usage must be disclosed.

---

### 5. System Architecture

#### 5.1 High‑Level Components
- **Frontend**: React 18 + TanStack Start (Vite, React Router, TanStack Query, Server Functions) + Tailwind CSS.  
- **Backend**: Go (latest stable) + Ent ORM. Exposes REST API under `/api/v1`.  
- **Database**: PostgreSQL 15+ with PostGIS extension.  
- **File Storage**: Local filesystem (development) or S3‑compatible storage (production, configurable).  
- **Reverse Proxy**: Nginx or similar for TLS termination and static file serving.

#### 5.2 Data Flow
1. Client sends HTTP request to backend.  
2. Backend authenticates via JWT middleware.  
3. Backend validates input and interacts with PostgreSQL via Ent.  
4. For file uploads, backend stores file and records metadata in DB.  
5. For entity extraction, backend runs extraction and stores entities.  
6. For PDF export, backend generates PDF and returns as binary.

---

### 6. Data Models (Ent Schemas)

#### 6.1 User
- `id` (UUID)  
- `email` (unique, index)  
- `password_hash`  
- `role` (`admin`, `analyst`, `viewer`)  
- `created_at`, `updated_at`  
- `is_active`

#### 6.2 Case
- `id` (UUID)  
- `title` (string, required)  
- `description` (text)  
- `status` (`open`, `closed`, `archived`)  
- `visibility` (`private`, `shared`, `public`)  
- `owner_id` (FK → User)  
- `tags` (M2M → Tag)  
- `created_at`, `updated_at`

#### 6.3 Evidence
- `id` (UUID)  
- `case_id` (FK → Case, index)  
- `type` (`text`, `url`, `image`, `video`, `document`, `raw`)  
- `title` (string)  
- `description` (text)  
- `source` (string)  
- `date` (timestamp with time zone, nullable)  
- `date_precision` (`exact`, `day`, `month`, `year`, `unknown`)  
- `location` (PostGIS geometry point, nullable)  
- `file_path` (string, nullable)  
- `content_text` (text, for extracted text or note content)  
- `metadata` (JSONB, for custom fields)  
- `created_by` (FK → User)  
- `created_at`, `updated_at`

#### 6.4 Entity
- `id` (UUID)  
- `type` (`person`, `organization`, `location`, `other`)  
- `name` (string)  
- `evidence_id` (FK → Evidence, on delete cascade)  
- `case_id` (denormalised for faster queries)  
- `created_at`

#### 6.5 Comment
- `id` (UUID)  
- `user_id` (FK → User)  
- `case_id` (FK → Case, nullable)  
- `evidence_id` (FK → Evidence, nullable)  
- `content` (text)  
- `created_at`, `updated_at`

#### 6.6 Tag
- `id` (UUID)  
- `name` (string, unique)  
- Many‑to‑many relationships with Case and Evidence.

#### 6.7 CaseMember
- `id` (UUID)  
- `case_id` (FK)  
- `user_id` (FK)  
- `role` (`viewer`, `editor`, `owner`)  
- `created_at`

#### 6.8 ActivityLog
- `id` (UUID)  
- `user_id` (FK)  
- `case_id` (FK)  
- `action` (string: `create`, `update`, `delete`, `comment`, etc.)  
- `details` (JSONB)  
- `created_at`

---

### 7. API Requirements

#### 7.1 General
- [x] All endpoints under `/api/v1`.  
- [x] Authentication via `Authorization: Bearer <token>` header.  
- [x] JSON request/response bodies.  
- [x] Use proper HTTP status codes (200, 201, 400, 401, 403, 404, 500).  
- [x] Pagination via `page` and `page_size` query parameters; default 20, max 100.  
- [x] Error response format: `{ "error": { "code": "string", "message": "human readable" } }`.

#### 7.2 Endpoint List (Summary)
- [x] `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`  
- [x] `GET /users/me`  
- [x] `GET /cases` (with filters)  
- [x] `POST /cases`  
- [x] `GET /cases/{id}`  
- [x] `PUT /cases/{id}`  
- [x] `DELETE /cases/{id}`  
- [x] `POST /cases/{id}/members`  
- [x] `DELETE /cases/{id}/members/{userId}`  
- [x] `GET /cases/{id}/evidence` (with filters)  
- [x] `POST /cases/{id}/evidence`  
- [x] `GET /evidence/{id}`  
- [x] `PUT /evidence/{id}`  
- [x] `DELETE /evidence/{id}`  
- [x] `GET /cases/{id}/evidence/geojson`  
- [x] `GET /cases/{id}/timeline`  
- [x] `POST /cases/{id}/comments`  
- [x] `GET /cases/{id}/comments`  
- [x] `POST /evidence/{id}/extract-entities`  
- [x] `GET /search?q=...&...`  
- [x] `GET /export/case/{id}/pdf`  
- [x] `GET /export/case/{id}/csv`

---

### 8. UI/UX Requirements

#### 8.1 Pages
- [x] **Login / Register**  
- [x] **Dashboard**  
- [x] **Case List**  
- [x] **Case Detail** (with tabs: Overview, Evidence, Map, Timeline, Entities, Comments, Members, Export)  
- [x] **Evidence Form** (modal or separate page)  
- [ ] **User Profile / Settings**

#### 8.2 Key Components
- [x] **Leaflet Map**: Displays markers, popups with evidence summary, draw controls (polygon, circle), zoom to fit.  
- [x] **Timeline**: Horizontal/vertical timeline with events, filtering.  
- [ ] **Entity Tag Cloud**: List of extracted entities with counts; click to filter evidence.  
- [x] **Comment Thread**: Simple list with reply and mention support.  
- [ ] **Data Table**: Sortable, paginated table for evidence list.

---

### 9. Security Requirements

- [ ] Use HTTPS everywhere.  
- [x] Implement CSRF protection for cookie‑based sessions (if used); JWT in localStorage mitigates some CSRF but must be handled carefully.  
- [ ] Validate file uploads (MIME type, extension, size).  
- [x] Use parameterised queries via Ent to prevent SQL injection.  
- [x] Sanitise all user‑generated content rendered in the frontend (React escapes by default).  
- [x] Rate limit authentication endpoints to prevent brute force.  
- [ ] Log security events (failed logins, permission denials).  
- [x] Store only necessary personal data.

---

### 10. Testing Requirements

- [ ] **Unit Tests**: Go handlers using `net/http/httptest` and mocked services; React components with Vitest + Testing Library.  
- [ ] **Integration Tests**: Database operations with a test PostgreSQL instance; API endpoint tests.  
- [ ] **End‑to‑End Tests**: Optional, using Playwright or Cypress for critical user flows (create case, add evidence, view map).  
- [ ] **Test Coverage**: Minimum 70% for backend, 50% for frontend critical components.

---

### 11. Documentation Requirements

- [x] **README.md**: Setup instructions (local development, Docker), environment variables, architecture overview.  
- [x] **API Documentation**: OpenAPI (Swagger) specification or a simple markdown file listing endpoints.  
- [ ] **User Guide**: Basic how‑to for the main features.

---

### 12. Assumptions & Constraints

- [x] The system will only process publicly available information; users are responsible for compliance with applicable laws.  
- [x] The initial version will not include real‑time collaboration (e.g., WebSockets); page refresh is acceptable.  
- [x] Entity extraction will use simple rules/regex; advanced NLP is out of scope for v1.  
- [x] File storage is local filesystem by default, but the interface must support swapping to cloud storage.  
- [x] The platform will be single‑tenant (one organisation per deployment) in the first release.

---

**End of Requirements Document**