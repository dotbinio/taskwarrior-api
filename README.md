# Taskwarrior API Server

> **⚠️ UNDER CONSTRUCTION**: This project is currently in active development and is **not ready for production use**. APIs may change without notice. Use at your own risk.

A headless REST API server for [Taskwarrior](https://taskwarrior.org/), providing a clean HTTP interface to interact with your tasks programmatically.

## Overview

This server acts as a bridge between Taskwarrior's powerful CLI and modern applications, allowing you to:

- Build web, mobile, or desktop UIs for Taskwarrior
- Integrate Taskwarrior with other tools and services
- Access your tasks from anywhere via HTTP
- Keep Taskwarrior as the single source of truth (no database duplication)

### Key Features

- **CLI-Only Integration**: Uses Taskwarrior CLI exclusively - no direct file manipulation
- **RESTful API**: Clean, predictable HTTP endpoints
- **Token Authentication**: Simple bearer token authentication
- **Sync-Friendly**: Compatible with Taskwarrior sync or file syncing (Syncthing, etc.)
- **No State Duplication**: All data lives in Taskwarrior

## Installation

### Prerequisites

- Go 1.21 or higher
- Taskwarrior installed and configured (`task` command available)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/dotbinio/taskwarrior-api.git
cd taskwarrior-api

# Install dependencies
make install

# Set required environment variable
export TW_API_TOKENS="your-secret-token"

# Build and run
make build
./bin/taskwarrior-api
```

The server will start on `http://localhost:8080`

**Access Swagger UI**: Open `http://localhost:8080/swagger/index.html` in your browser

## Embedded Example UI

A read-only web UI is available for viewing tasks and reports. This is intended as an example/development tool and is **enabled by default**.

### Accessing the UI

Simply start the server and open `http://localhost:8080/` in your browser.

### Disabling the UI

If you want to disable the UI (e.g., for production):

```bash
export TW_API_ENABLE_UI=false
./bin/taskwarrior-api
```

### Features

- Read-only interface for viewing tasks
- Dynamic report viewer using Taskwarrior's report configurations
- Displays tasks using each report's configured columns and labels
- Clean, minimal styling with Pico CSS
- No build step required (all dependencies via CDN)

### Security Notes

- **Enabled by default** - Can be disabled via `TW_API_ENABLE_UI=false`
- Requires valid API token (stored in browser session only)
- Read-only - cannot create, modify, or delete tasks
- Recommended for development and internal use
- Consider using a reverse proxy with additional authentication for external access

### Building from Source

```bash
# Clone the repository
git clone https://github.com/dotbinio/taskwarrior-api.git
cd taskwarrior-api

# Install dependencies
make install

# Build the binary
make build

# The binary will be available at ./bin/taskwarrior-api
```

### Running

```bash
# Run directly with Go
make run

# Or run the built binary
./bin/taskwarrior-api
```

## Docker Deployment

### Build and Push

```bash
# Build
docker build -t taskwarrior-api:latest .

# Tag for your registry
docker tag taskwarrior-api:latest your-registry/taskwarrior-api:latest

# Push
docker push your-registry/taskwarrior-api:latest
```

Or use the Makefile:

```bash
make docker-build
make docker-push  # Shows push instructions
```

## Kubernetes Deployment

Complete Kubernetes manifests are provided in `k8s/deployment.yaml`.

### Quick Deploy

```bash
# 1. Create secret with your tokens
kubectl create secret generic taskwarrior-api-secret \
  --from-literal=tokens="your-token-1,token-2" \
  -n taskwarrior

# 2. Update image in k8s/deployment.yaml
# 3. Deploy
kubectl apply -f k8s/deployment.yaml

# 4. Port forward to test
kubectl port-forward -n taskwarrior svc/taskwarrior-api 8080:8080
```

Or use the Makefile:

```bash
make k8s-deploy
make k8s-port-forward
```

See [KUBERNETES.md](KUBERNETES.md) for complete deployment guide including:
- Image registry setup
- Ingress configuration
- Monitoring and scaling
- Data persistence options
- Troubleshooting

## Configuration

The server is configured entirely through environment variables.

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TW_API_TOKENS` | Comma-separated list of valid auth tokens | `token1,token2,token3` |

### Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TW_API_HOST` | Server host address | `0.0.0.0` |
| `TW_API_PORT` | Server port | `8080` |
| `TW_API_ENABLE_UI` | Enable embedded example UI | `true` |
| `TW_DATA_LOCATION` | Path to Taskwarrior data directory | `~/.task` |
| `TW_TASKRC_LOCATION` | Path to Taskwarrior taskrc file | `~/.taskrc` |
| `TW_API_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `TW_API_CORS_ENABLED` | Enable CORS | `true` |
| `TW_API_CORS_ORIGINS` | Comma-separated list of allowed origins | `http://localhost:3000` |

### Example Configuration

```bash
export TW_API_TOKENS="my-secret-token-123,another-token-456"
export TW_API_PORT=8080
export TW_DATA_LOCATION="~/.task"
export TW_TASKRC_LOCATION="~/.taskrc"
export TW_API_LOG_LEVEL="info"
export TW_API_CORS_ORIGINS="http://localhost:3000,https://mytasks.example.com"

./bin/taskwarrior-api
```

Or create a `.env` file (see `env.example` for template) and source it:

```bash
cp env.example .env
# Edit .env with your values
source .env
./bin/taskwarrior-api
```

## API Documentation

All API endpoints require authentication via Bearer token.

### Interactive API Documentation (Swagger UI)

Once the server is running, you can access the interactive Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

This provides a web interface to:
- Browse all available endpoints
- View request/response schemas
- Test API calls directly from your browser
- See example requests and responses

### Authentication

Include the token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer your-token-here" http://localhost:8080/api/v1/tasks
```

### Endpoints

#### Health Check

**No authentication required**

```
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "taskwarrior-api"
}
```

---

### Tasks

#### List Tasks

```
GET /api/v1/tasks
```

Query parameters:
- `status` (default: `pending`) - Filter by status (pending, completed, deleted, waiting)
- `project` - Filter by project name
- `tags` - Filter by tags (can be specified multiple times)

Example:
```bash
curl -H "Authorization: Bearer token" \
  "http://localhost:8080/api/v1/tasks?status=pending&project=work"
```

Response:
```json
{
  "tasks": [
    {
      "uuid": "a360fc44-315c-4366-b70c-ea7e7520b749",
      "description": "Complete project documentation",
      "status": "pending",
      "project": "work",
      "tags": ["documentation"],
      "urgency": 8.9,
      "entry": "2026-01-01T10:00:00Z"
    }
  ],
  "count": 1
}
```

#### Get Task

```
GET /api/v1/tasks/:uuid
```

Example:
```bash
curl -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/tasks/a360fc44-315c-4366-b70c-ea7e7520b749
```

#### Create Task

```
POST /api/v1/tasks
```

Request body:
```json
{
  "description": "New task",
  "project": "work",
  "tags": ["important", "urgent"],
  "priority": "H",
  "due": "2026-01-15T00:00:00Z"
}
```

Required fields:
- `description` (string)

Optional fields:
- `project` (string)
- `tags` (array of strings)
- `priority` (string: H, M, L)
- `due` (ISO 8601 datetime)
- `wait` (ISO 8601 datetime)
- `scheduled` (ISO 8601 datetime)
- `depends` (array of UUIDs)
- `recur` (string: daily, weekly, monthly, etc.)

Example:
```bash
curl -X POST -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"description":"Write API documentation","project":"docs","tags":["writing"]}' \
  http://localhost:8080/api/v1/tasks
```

#### Update Task

```
PATCH /api/v1/tasks/:uuid
```

Request body (all fields optional):
```json
{
  "description": "Updated description",
  "project": "work",
  "priority": "H",
  "due": "2026-01-20T00:00:00Z"
}
```

Example:
```bash
curl -X PATCH -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"priority":"H","project":"urgent"}' \
  http://localhost:8080/api/v1/tasks/a360fc44-315c-4366-b70c-ea7e7520b749
```

#### Delete Task

```
DELETE /api/v1/tasks/:uuid
```

Example:
```bash
curl -X DELETE -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/tasks/a360fc44-315c-4366-b70c-ea7e7520b749
```

#### Mark Task as Done

```
POST /api/v1/tasks/:uuid/done
```

Example:
```bash
curl -X POST -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/tasks/a360fc44-315c-4366-b70c-ea7e7520b749/done
```

#### Start Task

```
POST /api/v1/tasks/:uuid/start
```

Example:
```bash
curl -X POST -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/tasks/a360fc44-315c-4366-b70c-ea7e7520b749/start
```

#### Stop Task

```
POST /api/v1/tasks/:uuid/stop
```

Example:
```bash
curl -X POST -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/tasks/a360fc44-315c-4366-b70c-ea7e7520b749/stop
```

---

### Reports

#### Next Report

Get pending tasks sorted by urgency.

```
GET /api/v1/reports/next
```

#### Active Report

Get currently active (started) tasks.

```
GET /api/v1/reports/active
```

#### Completed Report

Get completed tasks.

```
GET /api/v1/reports/completed
```

#### Waiting Report

Get tasks in waiting state.

```
GET /api/v1/reports/waiting
```

#### All Report

Get all tasks regardless of status.

```
GET /api/v1/reports/all
```

Example:
```bash
curl -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/reports/next
```

---

### Projects

#### List Projects

Get all projects with task counts.

```
GET /api/v1/projects
```

Response:
```json
{
  "projects": [
    {
      "name": "work",
      "count": 5
    },
    {
      "name": "personal",
      "count": 3
    }
  ],
  "count": 2
}
```

#### Get Project Tasks

Get all tasks for a specific project.

```
GET /api/v1/projects/:name/tasks
```

Example:
```bash
curl -H "Authorization: Bearer token" \
  http://localhost:8080/api/v1/projects/work/tasks
```

---

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "Human readable error message",
  "code": "ERROR_CODE"
}
```

Common error codes:
- `MISSING_AUTH_HEADER` - No Authorization header provided
- `INVALID_AUTH_FORMAT` - Authorization header format is incorrect
- `INVALID_TOKEN` - Token is not valid
- `INVALID_UUID` - Task UUID format is invalid
- `TASK_NOT_FOUND` - Task with given UUID doesn't exist
- `INVALID_REQUEST` - Request body is malformed

HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `500` - Internal Server Error

## Development

### Running Tests

```bash
make test
```

### Running with Hot Reload

Install [air](https://github.com/air-verse/air):

```bash
go install github.com/air-verse/air@latest
```

Then run:

```bash
make dev
```

### Code Formatting

```bash
make fmt
```

## Architecture

```
┌─────────────┐
│   Client    │
│ (Web/Mobile)│
└──────┬──────┘
       │ HTTP/REST
       │
┌──────▼──────────────┐
│   API Server (Go)   │
│  - Auth Middleware  │
│  - Route Handlers   │
└──────┬──────────────┘
       │ CLI Commands
       │
┌──────▼──────────────┐
│  Taskwarrior CLI    │
│   - task export     │
│   - task add        │
│   - task modify     │
└──────┬──────────────┘
       │ File I/O
       │
┌──────▼──────────────┐
│   .task/ Files      │
│  (Single Source of  │
│      Truth)         │
└─────────────────────┘
```

### Design Principles

1. **Taskwarrior is the backend** - No database, no state duplication
2. **CLI-only integration** - Never touch `.task/` files directly
3. **Strict separation** - Server and UI are completely independent
4. **Safe by default** - No arbitrary command execution
5. **Local-first** - Works with existing Taskwarrior + sync setups

## Security Considerations

- **Token Security**: Use strong, randomly generated tokens. Keep them secret.
- **HTTPS**: Always use HTTPS in production to protect tokens in transit.
- **Network Access**: Consider running behind a reverse proxy (nginx, Caddy) with additional security layers.
- **Local Use**: For maximum security, bind to `127.0.0.1` and use SSH tunneling for remote access.

## Use Cases

- **Web UI**: Build a modern web interface for Taskwarrior
- **Mobile Apps**: Create native mobile apps with offline sync
- **Automation**: Integrate with tools like Zapier, IFTTT, or custom scripts
- **Team Dashboards**: Display team tasks on shared screens
- **Voice Assistants**: Add tasks via Alexa, Google Home, etc.

## Production Deployment

This project includes production-ready deployment configurations:

- **Docker**: Multi-stage build with security best practices
- **Docker Compose**: Ready-to-use orchestration
- **CI/CD**: GitHub Actions workflows for testing and releases
- **Systemd**: Service configuration for Linux servers
- **Reverse Proxy**: Examples for Nginx and Caddy

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment guides, security considerations, monitoring, and troubleshooting.

### Quick Deploy with Docker

```bash
export TW_API_TOKENS="$(openssl rand -hex 32)"
docker-compose up -d
```

Access at: `http://localhost:8080/swagger/index.html`

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Taskwarrior](https://taskwarrior.org/) - The amazing task management tool this API wraps
