# URL Shortener Service

A production-grade URL shortening service built with Go 1.25, featuring clean architecture, PostgreSQL persistence, and comprehensive API endpoints.

## Features

- ✅ RESTful API for URL shortening operations
- ✅ PostgreSQL persistence with connection pooling
- ✅ Custom short codes support
- ✅ URL expiration with automatic cleanup
- ✅ Access count tracking
- ✅ Health check endpoint
- ✅ Structured logging with slog
- ✅ Graceful shutdown
- ✅ Context-aware request handling
- ✅ Environment-based configuration
- ✅ Docker support with multi-stage builds
- ✅ Production-ready error handling

## Architecture

```
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration loading and validation
│   ├── domain/          # Domain models and errors
│   ├── handler/         # HTTP handlers and routing
│   ├── repository/      # Database operations
│   ├── service/         # Business logic
│   └── storage/         # Database connection management
└── migrations/          # SQL migration files
```

## Prerequisites

- Go 1.25 or later
- PostgreSQL 12 or later
- Docker and Docker Compose (optional)

## Installation

### Local Development

1. Clone the repository:
```bash
git clone <repository-url>
cd url-shortener
```

2. Install dependencies:
```bash
go mod download
```

3. Set up PostgreSQL database:
```bash
createdb urlshortener
psql urlshortener < migrations/001_create_urls_table.up.sql
```

4. Create `.env` file from example:
```bash
cp .env.example .env
# Edit .env with your configuration
```

5. Run the service:
```bash
go run cmd/server/main.go
```

### Docker Deployment

1. Start the entire stack (PostgreSQL + Application):
```bash
docker compose up --build -d
```

2. Check logs:
```bash
docker compose logs -f app
```

3. Stop the stack:
```bash
docker compose down
```

## Configuration

Configuration can be provided via:
1. Environment variables (recommended for production)
2. `.env` file (for local development)
3. YAML configuration file (optional, set `CONFIG_FILE` env var)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `SERVER_READ_TIMEOUT` | HTTP read timeout | `10s` |
| `SERVER_WRITE_TIMEOUT` | HTTP write timeout | `10s` |
| `SERVER_IDLE_TIMEOUT` | HTTP idle timeout | `60s` |
| `SERVER_SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `30s` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | (required) |
| `DB_NAME` | Database name | `urlshortener` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `DB_MAX_OPEN_CONNS` | Max open connections | `25` |
| `DB_MAX_IDLE_CONNS` | Max idle connections | `5` |
| `DB_CONN_MAX_LIFETIME` | Connection max lifetime | `5m` |
| `DB_CONN_MAX_IDLE_TIME` | Connection max idle time | `5m` |
| `URL_SHORT_CODE_LENGTH` | Length of generated short codes | `7` |
| `URL_DEFAULT_TTL` | Default URL TTL (0 = no expiration) | `0` |
| `URL_BASE_URL` | Base URL for shortened links | `http://localhost:8080` |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json, text) | `json` |

## API Documentation

### Health Check

**GET** `/health`

Check service health status.

**Response:**
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Create Short URL

**POST** `/api/urls`

Create a new shortened URL.

**Request Body:**
```json
{
  "url": "https://example.com/very/long/url",
  "custom_code": "mycode",
  "ttl": 3600
}
```

**Fields:**
- `url` (required): The URL to shorten
- `custom_code` (optional): Custom short code (3-20 alphanumeric characters)
- `ttl` (optional): Time-to-live in seconds (0 = no expiration)

**Response (201):**
```json
{
  "id": 1,
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2026-01-29T10:00:00Z",
  "expires_at": "2026-01-29T11:00:00Z"
}
```

### Redirect to Original URL

**GET** `/{shortCode}`

Redirect to the original URL. Increments access count.

**Response:** HTTP 301 redirect to original URL

### Get URL Metadata

**GET** `/api/urls/{shortCode}`

Retrieve URL metadata without redirecting or incrementing access count.

**Response (200):**
```json
{
  "id": 1,
  "short_code": "abc123",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2026-01-29T10:00:00Z",
  "expires_at": "2026-01-29T11:00:00Z",
  "access_count": 42,
  "last_accessed": "2026-01-29T10:30:00Z"
}
```

### List URLs

**GET** `/api/urls?limit=20&offset=0`

Retrieve a paginated list of URLs.

**Query Parameters:**
- `limit` (optional): Number of results per page (default: 20, max: 100)
- `offset` (optional): Number of results to skip (default: 0)

**Response (200):**
```json
{
  "urls": [
    {
      "id": 1,
      "short_code": "abc123",
      "original_url": "https://example.com/url",
      "created_at": "2026-01-29T10:00:00Z",
      "expires_at": null,
      "access_count": 42,
      "last_accessed": "2026-01-29T10:30:00Z"
    }
  ],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

### Delete URL

**DELETE** `/api/urls/{shortCode}`

Delete a shortened URL.

**Response:** HTTP 204 No Content

## Error Responses

All errors follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

**Common Error Codes:**
- `400 Bad Request`: Invalid input
- `404 Not Found`: URL not found
- `409 Conflict`: Short code already exists
- `410 Gone`: URL has expired
- `500 Internal Server Error`: Server error

## Usage Examples

### Using curl

Create a short URL:
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'
```

Create with custom code:
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "custom_code": "gh"}'
```

Create with expiration (1 hour):
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "ttl": 3600}'
```

Get URL metadata:
```bash
curl http://localhost:8080/api/urls/abc123
```

List URLs:
```bash
curl http://localhost:8080/api/urls?limit=10&offset=0
```

Test redirect:
```bash
curl -L http://localhost:8080/abc123
```

Delete URL:
```bash
curl -X DELETE http://localhost:8080/api/urls/abc123
```

## Database Schema

```sql
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    access_count BIGINT NOT NULL DEFAULT 0,
    last_accessed TIMESTAMP WITH TIME ZONE
);
```

**Indexes:**
- `idx_urls_short_code` on `short_code`
- `idx_urls_created_at` on `created_at DESC`
- `idx_urls_expires_at` on `expires_at` (partial index)
- `idx_urls_access_count` on `access_count DESC`

## Monitoring & Observability

### Logging

The service uses structured JSON logging with the following fields:
- `time`: Timestamp
- `level`: Log level (DEBUG, INFO, WARN, ERROR)
- `msg`: Log message
- Context-specific fields (method, path, duration, error, etc.)

### Health Checks

The `/health` endpoint provides:
- Overall service status
- Database connectivity status
- Detailed error information when unhealthy

### Metrics

Key operational metrics tracked:
- URL creation count
- Access count per URL
- Expired URL cleanup count
- Request duration and status codes

## Production Considerations

### Security

- ✅ No sensitive data in logs
- ✅ Input validation on all endpoints
- ✅ SQL injection prevention via parameterized queries
- ✅ Rate limiting recommended (implement via reverse proxy)
- ✅ HTTPS recommended (configure via reverse proxy)

### Performance

- ✅ Connection pooling configured
- ✅ Database indexes on frequently queried columns
- ✅ Efficient query patterns
- ✅ Request timeouts configured
- ✅ Graceful shutdown to prevent data loss

### Scalability

- Stateless design enables horizontal scaling
- Database connection pooling configured
- Consider adding Redis cache for frequently accessed URLs
- Use database read replicas for read-heavy workloads

### Reliability

- Automatic cleanup of expired URLs
- Health check endpoint for load balancer integration
- Graceful shutdown handling
- Database connection retry logic
- Structured error handling

## Development

### Running Tests

```bash
go test ./... -v -cover
```

### Building

```bash
go build -o bin/server ./cmd/server
```

### Building Docker Image

```bash
docker build -t url-shortener:latest .
```

## Troubleshooting

### Database Connection Issues

1. Verify PostgreSQL is running:
```bash
pg_isready -h localhost -p 5432
```

2. Check connection credentials in `.env`

3. Ensure database exists:
```bash
psql -l | grep urlshortener
```

### Port Already in Use

Change `SERVER_PORT` in `.env` or stop conflicting service:
```bash
lsof -ti:8080 | xargs kill -9
```

### Migration Errors

Re-run migrations:
```bash
psql urlshortener < migrations/001_create_urls_table.down.sql
psql urlshortener < migrations/001_create_urls_table.up.sql
```

## License

MIT License - see LICENSE file for details

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Support

For issues, questions, or contributions, please open an issue on GitHub.
