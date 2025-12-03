## Implementation Plan: GoDrive MVP

### Big Task: Deliver a production-ready GoDrive MVP

#### Phase 1 – Project Foundations _(Completed)_
- **Requirements Alignment**
  - MVP scope confirmed against `docs/project_overview.md`; all Phase 2–10 features remain out of scope for initial launch.
  - Success metrics defined (see Phase 1 outcomes table below) and circulated as acceptance criteria.
- **Architecture Validation**
  - Modular boundaries locked: `API Gateway (Gin)` ↔ `Metadata Service (Postgres)` ↔ `Object Store Adapter (MinIO)` with shared observability layer.
  - Docker Compose baseline decided with five services (`api`, `postgres`, `minio`, `prometheus`, `grafana`) plus supporting volumes and networks.
- **Development Environment**
  - Go toolchain standardized on Go 1.22 with `gofmt` and `golangci-lint` enforced via pre-commit hooks.
  - Shared `.env` template maintained in this document for consistent local setup.

##### Phase 1 Outcomes
- **Success Metrics**
  - API availability ≥ 99.5% measured weekly.
  - P95 upload latency ≤ 800 ms for files ≤ 10 MB.
  - Authentication success rate ≥ 99% excluding invalid credential attempts.
  - Storage quota enforcement accuracy ≥ 99.9% (no over-allocations).
- **Architecture Notes**
  - API service exposes REST endpoints on port 8080; depends on healthy connections to Postgres and MinIO.
  - Postgres (port 5432) stores users, buckets, file metadata, and usage records; initialized via migrations in later phases.
  - MinIO (port 9000/9001) hosts object storage with `godrive` default bucket and S3 credentials supplied via environment.
  - Prometheus scrapes API and MinIO metrics; Grafana dashboards visualize availability, latency, storage consumption.
- **Environment Standards**
  - Developer setup requires Docker Desktop, Go 1.22, `golangci-lint`, and `pre-commit`.
  - Git hooks configured through `pre-commit` to run `gofmt`, `golangci-lint`, and `go test ./...` (tests added in future phases).
- Secrets stored in `.env` (ignored by VCS); template variables provided below for local setup.

###### `.env` Template (Development)
```
GODRIVE_API_PORT=8080
GODRIVE_JWT_SECRET=change-me-to-a-32-byte-secret
GODRIVE_JWT_REFRESH_SECRET=change-me-to-a-64-byte-secret
GODRIVE_LOG_LEVEL=info

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=godrive
POSTGRES_USER=godrive_app
POSTGRES_PASSWORD=change-me
POSTGRES_SSL_MODE=disable

MINIO_ENDPOINT=minio:9000
MINIO_ROOT_USER=godrive
MINIO_ROOT_PASSWORD=change-me-strong-password
MINIO_BUCKET=godrive
MINIO_USE_SSL=false

GODRIVE_AUTH_ACCESS_TOKEN_TTL=15m
GODRIVE_AUTH_REFRESH_TOKEN_TTL=720h
GODRIVE_AUTH_BCRYPT_COST=12

PROMETHEUS_SCRAPE_INTERVAL=15s
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=change-me

ENVIRONMENT=development
ALLOW_CORS_ORIGINS=http://localhost:3000
```

#### Phase 2 – Core Platform Setup _(Completed)_
- **Repository Structure**
  - Go module initialized at `github.com/abduss/godrive` with Go 1.22 toolchain.
  - Application skeleton created: `cmd/api` entrypoint plus `internal/{auth,config,metrics,server,storage}` modules.
  - Makefile targets added for `tidy`, `fmt`, `lint`, `test`, `run`, and migrations, standardizing workflows.
- **Database & Migrations**
  - PostgreSQL schema authored for `users`, `buckets`, `files`, `refresh_tokens`, `usage_snapshots`, and `bucket_usage`.
  - Migration suite established under `migrations/` with CLI integration through the Makefile.
  - Seed administrator account included in initial migration for bootstrap access.
- **Object Storage Integration**
  - MinIO client factory implemented with automatic bucket provisioning.
  - Health endpoints expose readiness checks for PostgreSQL and MinIO; liveness endpoint added for orchestration.

#### Phase 3 – Authentication Module _(Completed)_
- **User Management**
  - Implemented `/v1/auth/register` and `/v1/auth/login` endpoints with Gin, input validation, and structured responses.
  - Passwords hashed via bcrypt using configurable cost; duplicate email constraint handled gracefully.
- **JWT Issuance & Refresh Tokens**
  - Access tokens (HS256) include user claims (`sub`, `email`, `is_admin`) and configurable TTLs.
  - Refresh tokens generated with cryptographically secure random bytes, HMAC-stored in PostgreSQL, and persisted via repository layer.
- **Testing & Security**
  - Added unit tests covering registration, duplicate detection, login success, and invalid credentials.
  - Seed admin user uses bcrypt-hashed password; auth configuration (secrets, TTLs, cost) loaded from environment with sane defaults.

#### Phase 4 – File Storage & Buckets _(Completed)_
- **Bucket Operations**
  - Added authenticated `/v1/buckets` CRUD endpoints with per-owner isolation and conflict handling.
  - Bucket deletions cascade to metadata and trigger MinIO cleanup prior to removal.
- **File Operations**
  - Implemented multipart uploads, downloads, deletes, and listings under `/v1/buckets/:bucketID/files`.
  - Metadata persisted in PostgreSQL with UUID-based object naming stored in MinIO; supports checksum validation and size enforcement (100 MB default).
- **Data Integrity & Usage**
  - SHA-256 checksums computed during upload and stored alongside metadata.
  - Bucket usage counters maintained transactionally; usage snapshots recorded on each change for aggregate tracking.

#### Phase 5 – Presigned URLs
- **Link Generation**
  - Provide signed URLs for uploads/downloads with configurable TTL.
  - Validate permissions and scope tokens to specific buckets/files.
- **Audit & Revocation**
  - Log presigned URL creations and accesses.
  - Support manual invalidation via admin endpoint.

#### Phase 6 – Usage Tracking & Quotas _(In Progress)_
- **Metrics Collection**
  - Usage snapshots now captured automatically whenever files are uploaded or deleted; bucket-level byte and file counters stay synchronized with MinIO.
  - Remaining work: request counting and scheduled reporting.
- **Quota Enforcement**
  - Baseline size limit enforced per upload (100 MB). API-level quota management still pending.
- **Reporting**
  - Reporting endpoints not yet implemented; planned for later milestone.

#### Phase 7 – Logging, Monitoring & Observability
- **Structured Logging**
  - Adopt logging library (zap/logrus) with request correlation IDs.
  - Define logging standards (levels, fields) across modules.
- **Metrics & Alerting**
  - Instrument API with Prometheus metrics (latency, errors, throughput).
  - Create Grafana dashboards and alert rules for critical metrics.
- **Tracing (Optional)**
  - Evaluate OpenTelemetry integration for distributed tracing readiness.

#### Phase 8 – Deployment & Operations
- **Docker Compose Stack**
  - Write Compose file bundling API, Postgres, MinIO, Prometheus, Grafana.
  - Provide volume mappings, health checks, and init scripts.
- **Configuration Management**
  - Document environment variables and secrets provisioning.
  - Prepare production-ready configuration profiles (dev/staging/prod).
- **CI/CD Pipeline**
  - Implement automated build, test, lint pipeline (GitHub Actions or similar).
  - Add security scans (dependency, image scanning).

#### Phase 9 – Quality Assurance & Hardening
- **Testing Strategy**
  - Expand unit, integration, and e2e tests covering auth, storage, quotas.
  - Include load tests for critical endpoints (upload/download).
- **Documentation**
  - Create API reference (Swagger/OpenAPI) and setup guides.
  - Write operational runbooks (backups, recovery, scaling procedures).
- **Launch Readiness**
  - Perform security audit (OWASP top ten checklist).
  - Run pre-launch checklist: performance benchmarks, monitoring validation.

#### Phase 10 – Roadmap Preparation
- **Future Enhancements**
  - Outline migration path to microservices (service boundaries, messaging).
  - Capture backlog items: OAuth2, file versioning, analytics dashboard, cloud deployment.
- **Feedback Loop**
  - Define process for user feedback, issue triage, and iterative release planning.

### Deliverables
- Running GoDrive MVP via Docker Compose with full feature set.
- Comprehensive documentation (`README`, API docs, ops runbooks).
- CI/CD pipelines with automated testing and observability dashboards.
- Prioritized backlog for post-MVP roadmap features.

