# Quick Start Guide

## Option 1: Run with Docker Compose (Recommended)

This is the easiest way to get started as it includes both the application and PostgreSQL database.

```bash
# 1. Copy environment configuration
cp .env.example .env

# 2. Start the entire stack
docker compose up --build -d

# 3. Check if service is running
curl http://localhost:8080/health

# 4. Run API tests
bash test-api.sh

# 5. Stop the stack
docker compose down
```

## Option 2: Run Locally (Requires PostgreSQL)

### Prerequisites
- PostgreSQL 12+ installed and running
- Go 1.25+ installed

### Steps

```bash
# 1. Create database
createdb urlshortener

# 2. Run migrations
psql urlshortener < migrations/001_create_urls_table.up.sql

# 3. Setup environment
cp .env.example .env
# Edit .env with your PostgreSQL credentials

# 4. Download dependencies
go mod download

# 5. Run the server
go run cmd/server/main.go
```

## Option 3: Use the Setup Script

```bash
# Automated setup with Docker
bash setup.sh
```

## Quick API Examples

### Create a short URL
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'
```

### Create with custom code
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://golang.org", "custom_code": "go"}'
```

### Get URL metadata
```bash
curl http://localhost:8080/api/urls/abc123
```

### Use the short URL (redirect)
```bash
curl -L http://localhost:8080/abc123
```

### List all URLs
```bash
curl http://localhost:8080/api/urls
```

### Delete a URL
```bash
curl -X DELETE http://localhost:8080/api/urls/abc123
```

## Using Make Commands

The project includes a Makefile with convenient commands:

```bash
make help          # Show all available commands
make build         # Build the binary
make run           # Run locally
make docker-up     # Start with Docker
make docker-down   # Stop Docker containers
make docker-logs   # View application logs
make test          # Run tests
make api-test      # Run full API test suite
```

## Troubleshooting

### Port 8080 already in use
```bash
# Change SERVER_PORT in .env
SERVER_PORT=8081
```

### Database connection error
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Verify .env credentials match your PostgreSQL setup
```

### Docker issues
```bash
# View logs
docker compose logs app

# Restart containers
docker compose restart

# Full cleanup
docker compose down -v
docker compose up --build -d
```

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Check [API Documentation](README.md#api-documentation) for all endpoints
- Review [Configuration](README.md#configuration) options
- Explore [Production Considerations](README.md#production-considerations)

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/          # Domain models
│   ├── handler/         # HTTP handlers
│   ├── repository/      # Database operations
│   ├── service/         # Business logic
│   └── storage/         # Database connection
├── migrations/          # SQL migrations
├── docker-compose.yml   # Docker Compose configuration
├── Dockerfile           # Production Docker image
├── Makefile            # Development commands
└── README.md           # Full documentation
```

## Support

For detailed information, see [README.md](README.md).
For issues or questions, check the troubleshooting section above.
