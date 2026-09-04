# AGENTS.md

Project rules and quick reference for agents working in the `osint-app` repository. For the full Software Requirements Specification, see <ref_file file="/Users/akeelnazir/code/osint-app/REQUIREMENTS.md" />. For setup walkthrough, see <ref_file file="/Users/akeelnazir/code/osint-app/README.md" />.

## Project Overview

A full-stack Open Source Intelligence (OSINT) research platform inspired by Bellingcat methodologies. Analysts, journalists, and researchers create **cases**, ingest **evidence** (text, URL, image, video, document, raw data), enrich it with metadata, extract **entities**, and perform geospatial + temporal analysis with collaboration features (comments, mentions, activity logs). Lawful, open-source intelligence only. Single-tenant in v1.

## Tech Stack

- **Backend**: Go 1.27 + Ent ORM (`entgo.io/ent v0.14.6`), `chi/v5` router, `golang-jwt/v5`, `golang.org/x/crypto` (bcrypt), `golang.org/x/time` (rate limiting), `go-pdf/fpdf` (PDF export).
- **Frontend**: React 18 + TanStack Start (Vite + React Router + TanStack Query), TypeScript 5.7, Tailwind CSS 3.4, Leaflet 1.9 + `react-leaflet` 4.2 + `leaflet-draw`. Tests: Vitest 2 + Testing Library + jsdom + msw.
- **Database**: PostgreSQL 15+ with PostGIS extension (`postgis/postgis:15-3.4` image).
- **File Storage**: Local filesystem (dev) or S3-compatible (prod), switched via `FILE_STORAGE`.
- **Reverse Proxy**: Nginx or similar for TLS termination (production).

## Repo Layout

```
osint-app/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go        # HTTP server entry point
│   │   └── seed/main.go          # Database seeder
│   ├── internal/
│   │   ├── auth/                 # JWT + bcrypt (auth.go, auth_test.go)
│   │   ├── config/               # Env-based config
│   │   ├── ent/                  # Ent ORM generated code + schema/
│   │   ├── geo/                  # PostGIS helpers
│   │   ├── handler/              # HTTP handlers
│   │   ├── httperr/              # Structured API errors
│   │   ├── middleware/           # Auth, rate limit, security headers
│   │   ├── migrate/              # PostGIS + FTS migrations
│   │   ├── nlp/                  # Entity extraction (regex + gazetteer)
│   │   ├── server/               # Router wiring
│   │   └── storage/              # File storage (local/S3)
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── routes/               # File-based routes (TanStack Router)
│   │   │   ├── __root.tsx, _authed.tsx, login.tsx, register.tsx
│   │   │   └── _authed/ (index.tsx, search.tsx, cases/$caseId.tsx, cases/index.tsx)
│   │   ├── components/           # CommentsTab, EntitiesTab, EvidenceTab, ExportTab, MapTab, MembersTab, TimelineTab
│   │   ├── lib/                  # api.ts, api-services.ts, auth.tsx, mentions.ts, queryKeys.ts, types.ts
│   │   ├── router.tsx, routeTree.gen.ts, styles.css
│   ├── tests/                    # ExportTab, TimelineTab, mentions tests + setup.ts
│   ├── Dockerfile
│   ├── package.json
│   └── .env.example
├── docker-compose.yml            # db (PostGIS) + backend + frontend
├── README.md
├── REQUIREMENTS.md               # Full SRS
└── AGENTS.md                     # This file
```

## Common Commands

### Backend (run from `backend/`)

| Task | Command |
|------|---------|
| Run server | `go run ./cmd/server` |
| Seed database | `go run ./cmd/seed` |
| Run all tests | `go test ./...` |
| Regenerate Ent code | `GOTOOLCHAIN=go1.24.0 go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/ent/schema` |
| Download deps | `go mod download` |

> Ent's typechecker is incompatible with Go 1.27's stdlib layout. Always use `GOTOOLCHAIN=go1.24.0` when regenerating Ent. Normal builds/runs use the installed Go 1.27 toolchain.

### Frontend (run from `frontend/`)

| Task | Command |
|------|---------|
| Dev server | `npm run dev` |
| Build | `npm run build` |
| Preview build | `npm run preview` |
| Typecheck | `npm run typecheck` (`tsc --noEmit`) |
| Run tests once | `npm test` (`vitest run`) |
| Watch tests | `npm run test:watch` |

### Docker Compose (from repo root)

| Task | Command |
|------|---------|
| Start everything | `docker compose up -d` |
| Start DB only | `docker compose up db -d` |
| Seed via container | `docker compose exec backend /app/seed` |

### Access URLs (local dev)

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080/api/v1
- Health check: http://localhost:8080/api/v1/health

### Default Seed Credentials

- **Admin**: `admin@osint.local` / `adminpass123`
- **Analyst**: `analyst@osint.local` / `analystpass123`

## Environment Variables

Copy `backend/.env.example` → `backend/.env` and `frontend/.env.example` → `frontend/.env`. Never commit real `.env` files (already in `.gitignore`).

**Backend** (`backend/.env.example`):
- `DATABASE_URL` — PostgreSQL connection string (PostGIS must be enabled on the DB).
- `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` — JWT signing secrets. **Change in production.**
- `ACCESS_TOKEN_TTL` (default `15m`), `REFRESH_TOKEN_TTL` (default `168h`).
- `FILE_STORAGE` — `local` or `s3`.
- `LOCAL_STORAGE_PATH` (default `./uploads`).
- `S3_BUCKET`, `S3_REGION`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_ENDPOINT` — only used when `FILE_STORAGE=s3`.
- `APP_ENV` (`development` | `production`), `HTTP_ADDR` (default `:8080`).
- `CORS_ALLOWED_ORIGINS` — comma-separated allowed origins.
- `RATE_LIMIT_RPS` — requests per second per IP (default `20`).

**Frontend** (`frontend/.env.example`):
- `VITE_API_URL` — backend API base URL (default `http://localhost:8080/api/v1`).

## Architecture & Conventions

### API

- All endpoints under `/api/v1`. JSON request/response bodies.
- Auth via `Authorization: Bearer <token>` header.
- HTTP status codes: 200, 201, 400, 401, 403, 404, 500.
- Pagination via `page` and `page_size` query params; default 20, max 100.
- Error response shape: `{ "error": { "code": "string", "message": "human readable" } }` — use `internal/httperr` helpers.
- Endpoint list is documented in README.md §API Endpoints and REQUIREMENTS.md §7.2.

### Data Access

- **Always** use Ent ORM for DB access — parameterized queries prevent SQL injection. No raw SQL for user-controlled data.
- PostGIS `geom geography(Point,4326)` column on `evidence` for spatial queries (`ST_Within`, `ST_Distance`). Use `internal/geo` helpers.
- Full-text search uses a generated `tsvector` column on evidence (title, description, content). Fallback to `ILIKE` per FR-SRCH-1.
- Database migrations are versioned and reversible where possible (`internal/migrate`).
- Indexes on frequently queried fields: `case_id`, tags, dates, geospatial columns.

### Auth & RBAC

- bcrypt password hashing (never store or log plaintext passwords).
- JWT access (15 min) + refresh (7 days) tokens. Sign with strong secret (HS256) or RSA.
- Refresh token grants new access token without re-auth.
- Global roles: **admin**, **analyst**, **viewer**.
- Case-level roles: **owner**, **editor**, **viewer** (via `CaseMember`).
- A case must always have at least one owner.
- Enforce permissions in handlers/middleware — see REQUIREMENTS.md §2 for the full capability matrix.
- Rate-limit auth endpoints to prevent brute force; log failed logins + permission denials.

### File Uploads

- Validate MIME type, extension, and size (max 50 MB).
- Store via `internal/storage` (local filesystem or S3). DB stores the file path/URL + metadata.
- Deleting evidence must also remove associated entities, comments, and the file (if any).

### Frontend

- File-based routes under `src/routes` (TanStack Router). Authed routes nested under `_authed`.
- Reusable case-detail tabs in `src/components` (`EvidenceTab`, `MapTab`, `TimelineTab`, `EntitiesTab`, `CommentsTab`, `MembersTab`, `ExportTab`).
- API client + types + query keys in `src/lib`. Use TanStack Query for server state.
- Tailwind for styling. Responsive: desktop + tablet.
- React escapes output by default — do not inject raw user content via `dangerouslySetInnerHTML`.
- Inline form validation with clear error messages.

### Entity Extraction (v1)

- Simple regex + gazetteer NER in `internal/nlp` for persons, organizations, locations, emails, URLs, coordinates.
- Advanced NLP is out of scope for v1.
- Extracted entities stored in `Entity` table, linked to source evidence; denormalized `case_id` for faster queries.
- Users can manually add/remove entity links.

## Data Models (Ent Schemas)

Eight entities defined in `backend/internal/ent/schema/`:

| Entity | Key Fields |
|--------|------------|
| **User** | `id` (UUID), `email` (unique, index), `password_hash`, `role` (admin/analyst/viewer), `is_active`, timestamps |
| **CaseRecord** (table `cases`) | `id`, `title`, `description`, `status` (open/closed/archived), `visibility` (private/shared/public), `owner_id`, tags (M2M), timestamps |
| **Evidence** | `id`, `case_id` (FK, index), `type` (text/url/image/video/document/raw), `title`, `description`, `source`, `date`, `date_precision` (exact/day/month/year/unknown), `location` (PostGIS point), `file_path`, `content_text`, `metadata` (JSONB), `created_by`, timestamps |
| **Entity** | `id`, `type` (person/organization/location/other), `name`, `evidence_id` (FK, cascade), `case_id` (denormalized), `created_at` |
| **Comment** | `id`, `user_id`, `case_id` (nullable), `evidence_id` (nullable), `content`, timestamps |
| **Tag** | `id`, `name` (unique), M2M with Case and Evidence |
| **CaseMember** | `id`, `case_id`, `user_id`, `role` (viewer/editor/owner), `created_at` |
| **ActivityLog** | `id`, `user_id`, `case_id`, `action`, `details` (JSONB), `created_at` |

## Testing

- **Backend**: `go test ./...` from `backend/`. Covers NLP entity extraction, JWT auth, password hashing, config loading. Target ≥ 70% coverage.
- **Frontend**: `npm test` from `frontend/` (Vitest + Testing Library). Covers Timeline, Export, mention extraction. Target ≥ 50% coverage on critical components.
- Integration tests use a test PostgreSQL instance; E2E (Playwright/Cypress) is optional for critical flows (create case → add evidence → view map).

## Security Checklist

- HTTPS everywhere (TLS 1.2+).
- JWT signed with strong secret; never log passwords or tokens.
- Validate + sanitise all user input (XSS, SQL injection, path traversal).
- File upload validation (MIME, extension, size ≤ 50 MB).
- Rate limiting per IP (token bucket, `RATE_LIMIT_RPS`).
- CORS allowlist via `CORS_ALLOWED_ORIGINS` (frontend origin only).
- Security headers: `X-Content-Type-Options`, `X-Frame-Options`, HSTS, `Referrer-Policy`.
- Parameterized queries via Ent.
- Log security events (failed logins, permission denials).
- GDPR: support account deletion + personal data export; disclose cookie/localStorage usage.

## Out of Scope (v1)

- Real-time collaboration (WebSockets) — page refresh is acceptable.
- Advanced NLP for entity extraction (simple rules/regex only).
- Multi-tenant deployments (single-tenant per deployment in first release).

## Reference Documents

- <ref_file file="/Users/akeelnazir/code/osint-app/REQUIREMENTS.md" /> — full SRS (functional + non-functional requirements, API endpoint list, UI/UX, security, testing, compliance).
- <ref_file file="/Users/akeelnazir/code/osint-app/README.md" /> — setup instructions, architecture diagram, API endpoint tables, configuration reference.
