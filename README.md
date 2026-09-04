# OSINT Research App

A full-stack Open Source Intelligence (OSINT) research platform inspired by [Bellingcat](https://www.bellingcat.com/) methodologies. Collect, analyze, visualize, and collaborate on publicly available information.

## Tech Stack

- **Frontend**: React 18, TanStack Start (Vite + React Router + TanStack Query), TypeScript, Tailwind CSS, Leaflet
- **Backend**: Go 1.27, Ent ORM, chi router, JWT auth, gofpdf
- **Database**: PostgreSQL 15+ with PostGIS extension

## Features

1. **Authentication & Roles** — JWT (access + refresh), bcrypt, roles: admin/analyst/viewer
2. **Case Management** — CRUD, tags, sharing, public/private visibility, team members
3. **Evidence & Data Ingestion** — Text, URL, image, video, document, raw JSON; auto URL metadata scraping; file uploads
4. **Search & Filtering** — Full-text + ILIKE search, filter by type/date/tags/geo radius, pagination
5. **Geolocation & Mapping** — PostGIS-backed, Leaflet map with markers, popups, draw controls
6. **Timeline Analysis** — Chronological view of dated evidence, filter by date/type
7. **Entity Extraction** — Regex + gazetteer NER for people, organizations, locations, emails, URLs, coordinates
8. **Collaboration & Comments** — Case/evidence comments, @username mentions
9. **Export & Reporting** — PDF case reports (gofpdf), CSV evidence export
10. **Dashboard** — Stats, recent cases, activity feed

## Quick Start

### Prerequisites

- Go 1.27+
- Node.js 22+
- Docker & Docker Compose (for PostgreSQL)

### Option 1: Docker Compose (recommended)

```bash
# Start PostgreSQL with PostGIS + backend + frontend
docker compose up -d

# Run database migrations + seed data
docker compose exec backend /app/seed

# Access:
#   Frontend: http://localhost:5173
#   Backend:  http://localhost:8080
#   API:      http://localhost:8080/api/v1
```

### Option 2: Local Development

#### 1. Start PostgreSQL with PostGIS

```bash
docker compose up db -d
```

#### 2. Backend setup

```bash
cd backend
cp .env.example .env  # adjust if needed

# Install deps
go mod download

# Generate Ent code (requires Go 1.24 toolchain for codegen)
GOTOOLCHAIN=go1.24.0 go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/ent/schema

# Run migrations + start server
go run ./cmd/server

# In another terminal, seed the database
go run ./cmd/seed
```

#### 3. Frontend setup

```bash
cd frontend
cp .env.example .env  # adjust if needed
npm install
npm run dev
```

#### 4. Access

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080/api/v1
- Health check: http://localhost:8080/api/v1/health

### Default Seed Credentials

- **Admin**: admin@osint.local / adminpass123
- **Analyst**: analyst@osint.local / analystpass123

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login and receive tokens |
| POST | `/auth/refresh` | Refresh access token |

### User
| Method | Path | Description |
|--------|------|-------------|
| GET | `/users/me` | Get current user profile |

### Cases
| Method | Path | Description |
|--------|------|-------------|
| GET | `/cases` | List cases (filtered by access) |
| POST | `/cases` | Create a case |
| GET | `/cases/:id` | Get a case |
| PUT | `/cases/:id` | Update a case |
| DELETE | `/cases/:id` | Delete a case |
| GET | `/cases/:id/members` | List case members |
| POST | `/cases/:id/members` | Add a member |
| DELETE | `/cases/:id/members/:userId` | Remove a member |
| GET | `/cases/:id/timeline` | Get timeline events |
| GET | `/cases/:id/comments` | List case comments |
| POST | `/cases/:id/comments` | Add a comment |
| GET | `/cases/:id/activity` | List case activity |

### Evidence
| Method | Path | Description |
|--------|------|-------------|
| GET | `/cases/:id/evidence` | List evidence (with filters) |
| POST | `/cases/:id/evidence` | Create evidence (JSON or multipart) |
| GET | `/cases/:id/evidence/geojson` | Get GeoJSON for map |
| GET | `/cases/:id/evidence/:evidenceId` | Get evidence |
| PUT | `/cases/:id/evidence/:evidenceId` | Update evidence |
| DELETE | `/cases/:id/evidence/:evidenceId` | Delete evidence |
| POST | `/cases/:id/evidence/:evidenceId/entities` | Extract entities |
| GET | `/cases/:id/evidence/:evidenceId/entities` | List entities |
| GET | `/cases/:id/evidence/:evidenceId/comments` | List evidence comments |
| POST | `/cases/:id/evidence/:evidenceId/comments` | Add evidence comment |

### Search
| Method | Path | Description |
|--------|------|-------------|
| GET | `/search` | Global search (q, type, date_from, date_to, lat, lng, radius, tag) |

### Export
| Method | Path | Description |
|--------|------|-------------|
| GET | `/export/case/:id/pdf` | Download case PDF report |
| GET | `/export/case/:id/evidence.csv` | Download evidence CSV |

### Dashboard
| Method | Path | Description |
|--------|------|-------------|
| GET | `/dashboard` | Get dashboard stats |

## Architecture

```
osint-app/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go      # HTTP server entry point
│   │   └── seed/main.go        # Database seeder
│   ├── internal/
│   │   ├── auth/               # JWT + bcrypt
│   │   ├── config/             # Env-based config
│   │   ├── ent/                # Ent ORM (generated + schemas)
│   │   ├── geo/                # PostGIS helpers
│   │   ├── handler/            # HTTP handlers
│   │   ├── httperr/            # Structured API errors
│   │   ├── middleware/         # Auth, rate limit, security headers
│   │   ├── migrate/            # PostGIS + FTS migrations
│   │   ├── nlp/                # Entity extraction
│   │   ├── server/             # Router wiring
│   │   └── storage/            # File storage (local/S3)
│   ├── Dockerfile
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── routes/             # File-based routes (TanStack Router)
│   │   ├── components/         # Reusable components
│   │   ├── lib/                # API client, auth, types, query keys
│   │   └── styles.css          # Tailwind entry
│   ├── tests/                  # Vitest + Testing Library
│   ├── Dockerfile
│   └── .env.example
├── docker-compose.yml
└── README.md
```

## Database Schema

Ent schemas for 8 entities:
- **User** — email, username, password_hash, role
- **CaseRecord** (table: `cases`) — title, description, status, visibility, owner
- **Evidence** — type, title, content, file_path, lat/lng, geom (PostGIS), metadata
- **Entity** — type (person/org/location/other), value, confidence
- **Comment** — body, mentions, author, case/evidence link
- **Tag** — name (m2m with cases and evidence)
- **CaseMember** — case↔user join with role (owner/editor/viewer)
- **ActivityLog** — action, target, user, case

PostGIS adds a `geom geography(Point,4326)` column to evidence for spatial queries. Full-text search uses a generated `tsvector` column on evidence (title, description, content).

## Testing

### Backend
```bash
cd backend
go test ./...
```

Tests cover: NLP entity extraction, JWT auth, password hashing, config loading.

### Frontend
```bash
cd frontend
npm test
```

Tests cover: Timeline component, Export component, mention extraction utility.

## Security

- JWT access (15min) + refresh (7d) tokens
- bcrypt password hashing
- Per-IP rate limiting (token bucket)
- CORS with configurable allowed origins
- Security headers (X-Content-Type-Options, X-Frame-Options, HSTS, Referrer-Policy)
- Input validation on all endpoints
- Role-based access control (admin/analyst/viewer)
- Case-level access control (owner/editor/viewer members + public)

## Ent Code Generation

Ent's typechecker is incompatible with Go 1.27's stdlib layout. To regenerate Ent code:

```bash
cd backend
GOTOOLCHAIN=go1.24.0 go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/ent/schema
```

Normal builds use the installed Go 1.27 toolchain without issue.

## Configuration

See `backend/.env.example` and `frontend/.env.example` for all available environment variables.

Key variables:
- `DATABASE_URL` — PostgreSQL connection string
- `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` — JWT signing secrets
- `FILE_STORAGE` — `local` or `s3`
- `CORS_ALLOWED_ORIGINS` — Comma-separated allowed origins
- `RATE_LIMIT_RPS` — Requests per second per IP
- `VITE_API_URL` — Backend API URL (frontend)

## License

This is a demonstration project. Adapt and use according to your needs.
