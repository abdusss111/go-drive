# Project Information: GoDrive

## 1. Overview
**GoDrive** is a self-hosted, lightweight cloud storage backend built with Go. It mimics the core functionality of services like AWS S3 or Google Drive but in a self-contained, easy-to-deploy package.

The system relies on a **Modular Monolithic** architecture, ensuring clean separation of concerns while maintaining the simplicity of a single deployment unit for the API.

## 2. Technology Stack

### Core
- **Language**: Go 1.22+
- **Web Framework**: Gin (High-performance HTTP web framework)
- **Authentication**: JWT (JSON Web Tokens) with Access and Refresh token rotation.

### Infrastructure & Data
- **Database (Metadata)**: PostgreSQL (Stores user info, bucket metadata, file metadata, usage stats).
- **Object Storage**: MinIO (S3-compatible storage for actual file content).
- **Deployment**: Docker Compose for orchestration.

### Observability
- **Metrics**: Prometheus (Exposes metrics at `/metrics`).
- **Visualization**: Grafana (Pre-configured dashboards for API health and Storage stats).
- **Logging**: Zap (Structured logging with request correlation IDs).

## 3. Architecture

The application is structured as a **Modular Monolith**. Each core domain is encapsulated with its own Service, Repository, and HTTP Handlers.

### Directory Structure
- `cmd/api`: Main entry point. Bootstraps dependencies and starts the server.
- `internal/`:
  - **auth**: User registration, login, token generation/validation.
  - **bucket**: Logical grouping of files. Manages bucket creation and listing.
  - **file**: Core file operations (Upload, Download, Delete, List). Handles interaction with MinIO and Metadata DB.
  - **usage**: Tracks storage usage (bytes used, file counts) and enforces quotas.
  - **presigned**: Service for generating time-limited presigned URLs (GET/PUT) for secure, direct object access.
  - **storage**: Low-level clients for PostgreSQL and MinIO.
  - **server**: HTTP Router setup (Gin), Middleware configuration (CORS, Auth, Logger, Metrics).

### Data Flow
1. **Request**: Enters via `cmd/api` -> `server/router`.
2. **Middleware**: Logs request, adds Correlation ID, verifies Auth Token (if protected).
3. **Handler**: Parses request, calls Service.
4. **Service**:
   - Validates business rules (e.g., Quota checks).
   - Interacts with **Repositories** (PostgreSQL) for metadata.
   - Interacts with **ObjectStore** (MinIO) for binary data.
5. **Response**: Returns JSON response with standardized error handling.

## 4. Business Logic & Key Flows

### File Upload (`POST /v1/buckets/:bucketID/files`)
1. **Validation**: Checks if file exists in request.
2. **Quota Check**: Verifies if user has enough space (bytes) and file slots (count).
3. **Stream Processing**:
   - Generates a UUID for the file.
   - Streams content to MinIO using `PutObject`.
   - Simultaneously calculates SHA256 checksum (using `io.TeeReader`).
4. **Rollback Safety**: If MinIO upload succeeds but DB insert fails, the object is deleted from MinIO to ensure consistency.
5. **Metadata**: Stores file details (Name, Size, Content-Type, Checksum) in PostgreSQL.
6. **Usage Update**: Increments user/bucket usage stats.

### File Download
- Retrieves metadata from DB to verify ownership and get the real object path.
- Streams the file content directly from MinIO through the API to the client.

### Presigned URLs
- Allows generating temporary, secure URLs for direct file access or upload, bypassing the API server for data transfer which improves scalability.
- Validates ownership before generating the URL.

### Quotas & Usage
- **Tracking**: Maintained in real-time. Every upload/delete updates `usage` tables.
- **Enforcement**: Strict checks before allowing uploads. Limits defined per user (default 100MB per file, though customizable).

### Authentication
- **Register**: Creates user with bcrypt hashed password.
- **Login**: Returns `access_token` (Short TTL, e.g., 15m) and `refresh_token` (Long TTL, e.g., 30d).
- **Middleware**: Validates `Authorization: Bearer <token>` for protected routes.

## 5. Development & Operations

### Build & Run
- **Make**: Automation via `Makefile` (`make run`, `make migrate-up`, `make test`, `make fmt`).
- **Run**: `make setup` starts dependencies (Postgres, MinIO, Prometheus, Grafana) via Docker. `make run` starts the Go API.

### Configuration
Managed via environment variables (loaded from `.env`).
- `GODRIVE_API_PORT`, `GODRIVE_JWT_SECRET`
- `POSTGRES_HOST`, `MINIO_ENDPOINT`
- `GODRIVE_METRICS_PATH`

### Testing
- Integration and E2E tests located in `tests/`.
- Unit tests co-located with code (e.g., `service_test.go`).
