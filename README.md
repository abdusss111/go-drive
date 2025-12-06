# goDrive

**goDrive** - this is a cloud file storage service written in **Go**. It allows users to create folders, upload and download files via the REST API.


## Architecture

- **API**: Go + Gin, JWT authentication
- **Database**: PostgreSQL
- **File Storage**: MinIO (compatible with S3)
- **Metrics**: Prometheus
- **Containerization**: Docker+ Docker Compose

### Requirements

- Docker
- Docker Compose

### For production:

- Use SSL.
- More reliable JWT secrets.
- Enable POSTGRES_SSL_MODE=require.

### For developers
- The code is located in cmd/api and internal/
- Used by go 1.22
- For local development, MinIO and PostgreSQL run in containers
- To run locally without Docker, make sure that PostgreSQL and MinIO are running locally