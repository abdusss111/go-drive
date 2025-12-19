# GoDrive - Lightweight Cloud Storage Service

GoDrive is a self-hosted, lightweight cloud storage backend built with Go, providing secure file storage and management via a REST API.

## Features

- **User Authentication**: JWT-based authentication with access and refresh tokens
- **File Storage**: Upload, download, delete, and list files organized in buckets
- **Bucket Management**: Create, list, and manage storage buckets
- **Quota Enforcement**: Per-user storage quotas (bytes and file count limits)
- **Usage Tracking**: Real-time usage statistics and reporting
- **Structured Logging**: Request correlation IDs for tracing
- **Metrics**: Prometheus metrics for monitoring
- **Health Checks**: Liveness and readiness endpoints

## Architecture

GoDrive follows a modular monolithic architecture:

- **API Gateway**: Gin-based REST API (port 8080)
- **Metadata Storage**: PostgreSQL for users, buckets, files, and usage data
- **Object Storage**: MinIO (S3-compatible) for file objects
- **Observability**: Prometheus for metrics, Grafana for visualization

## Quick Start

### Prerequisites

- Docker Desktop
- Go 1.22+
- Make

### Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd drive
```

2. Create `.env` file (see `.env.example` for template):
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start services with Docker Compose:
```bash
docker-compose up --build
make migrate-up
```

4. Run the API server:
```bash
make run
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /v1/auth/register` - Register a new user
- `POST /v1/auth/login` - Login and get access token

### Buckets
- `POST /v1/buckets` - Create a new bucket
- `GET /v1/buckets` - List user's buckets
- `GET /v1/buckets/:bucketID` - Get bucket details
- `DELETE /v1/buckets/:bucketID` - Delete a bucket

### Files
- `POST /v1/buckets/:bucketID/files` - Upload a file
- `GET /v1/buckets/:bucketID/files` - List files in bucket
- `GET /v1/buckets/:bucketID/files/:fileID/download` - Download a file
- `DELETE /v1/buckets/:bucketID/files/:fileID` - Delete a file

### Usage
- `GET /v1/usage` - Get user's total usage statistics
- `GET /v1/buckets/:bucketID/usage` - Get bucket usage statistics

### Health & Metrics
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

## Configuration

### Environment Variables

#### Server
- `GODRIVE_API_PORT` - API server port (default: 8080)
- `GODRIVE_API_HOST` - API server host (default: 0.0.0.0)
- `GODRIVE_LOG_LEVEL` - Log level: debug, info, warn, error (default: info)

#### Database
- `POSTGRES_HOST` - PostgreSQL host (default: localhost)
- `POSTGRES_PORT` - PostgreSQL port (default: 5432)
- `POSTGRES_USER` - Database user (default: godrive_app)
- `POSTGRES_PASSWORD` - Database password
- `POSTGRES_DB` - Database name (default: godrive)
- `POSTGRES_SSL_MODE` - SSL mode (default: disable)

#### MinIO
- `MINIO_ENDPOINT` - MinIO endpoint (default: localhost:9000)
- `MINIO_ROOT_USER` - MinIO access key
- `MINIO_ROOT_PASSWORD` - MinIO secret key
- `MINIO_BUCKET` - Default bucket name (default: godrive)
- `MINIO_USE_SSL` - Use SSL (default: false)

#### Authentication
- `GODRIVE_JWT_SECRET` - JWT access token secret (32+ bytes)
- `GODRIVE_JWT_REFRESH_SECRET` - JWT refresh token secret (64+ bytes)
- `GODRIVE_AUTH_ACCESS_TOKEN_TTL` - Access token TTL (default: 15m)
- `GODRIVE_AUTH_REFRESH_TOKEN_TTL` - Refresh token TTL (default: 720h)
- `GODRIVE_AUTH_BCRYPT_COST` - Bcrypt cost (default: 12)

#### Monitoring
- `GODRIVE_METRICS_PATH` - Prometheus metrics path (default: /metrics)
- `GRAFANA_ADMIN_USER` - Grafana admin username (default: admin)
- `GRAFANA_ADMIN_PASSWORD` - Grafana admin password

## Development

### Running Tests
```bash
make test
```

### Code Formatting
```bash
make fmt
```

### Linting
```bash
make lint
```

### Database Migrations
```bash
# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down
```

## Docker Services

- **API**: `http://localhost:8080`
- **PostgreSQL**: `localhost:5433`
- **MinIO API**: `http://localhost:9000`
- **MinIO Console**: `http://localhost:9001`
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000`

## Project Structure

```
drive/
├── cmd/api/          # Application entrypoint
├── internal/         # Internal packages
│   ├── auth/        # Authentication module
│   ├── bucket/      # Bucket management
│   ├── file/        # File operations
│   ├── usage/       # Usage tracking and quotas
│   ├── logger/      # Structured logging
│   ├── metrics/     # Prometheus metrics
│   ├── server/      # HTTP server setup
│   └── storage/     # Database and MinIO clients
├── migrations/      # Database migrations
├── tests/           # Integration and E2E tests
└── docs/            # Documentation
```

## License

[Add your license here]

