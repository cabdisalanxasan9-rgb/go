# DefenseKit

Defensive cybersecurity toolkit written in Go with CLI scanners, plugin-based web API, and live dashboard updates.

> Warning: Educational and defensive use only. Do not scan systems without explicit authorization.

## Features

- CLI scanning tools: HTTP status/time, TCP ports, banner grabbing, subdomains, DNS, SSL, password entropy, latency.
- Worker pool concurrency with configurable `threads`, `timeout`, and rate limiting.
- Professional CLI flags with verbose/silent modes, JSON/TXT output, and log file support.
- REST API returning JSON for scans and plugin runs.
- Web dashboard with plugin execution and auto-refresh live results.
- Modular plugin system for adding new modules quickly.
- Docker support and CI workflow for production pipelines.

## Project Structure

```
defensekit/
├── cmd/
│   └── main.go
├── internal/
│   ├── scanner/          # HTTP, port, banner, DNS, SSL, latency
│   ├── subdomain/
│   ├── password/
│   ├── worker/
│   ├── output/
│   ├── logger/
│   └── web/
│       ├── api.go
│       ├── dashboard.go
│       ├── plugins/
│       │   └── example_plugin.go
│       └── templates/
│           ├── dashboard.html
│           └── live_updates.html
├── docker/
│   └── Dockerfile
├── .github/workflows/
│   └── ci.yml
├── go.mod
└── README.md
```

## CLI Usage

```bash
go run ./cmd -mode all -target scanme.nmap.org -start 1 -end 1000 -threads 200 -timeout 1 -rate 100 -output result.json -format json
go run ./cmd -mode http -target example.com
go run ./cmd -mode password -password "StrongP@ssw0rd123"
go run ./cmd -serve -addr :8080
go run ./cmd -healthcheck -health-url http://127.0.0.1:8080/api/health
```

## Key Flags

- `-mode`: `http|subdomains|portscan|password|dns|ssl|latency|all`
- `-target`: host/domain target
- `-threads`: worker count
- `-timeout`: timeout seconds
- `-rate`: rate limit per second
- `-verbose` / `-silent`
- `-format`: `json|txt`
- `-output`: output path
- `-log`: log file path
- `-healthcheck`: run healthcheck and exit (`0` healthy, non-zero unhealthy)
- `-health-url`: health endpoint URL

## Environment Variables

All primary flags can be set via environment variables (flags still override env values):

- `DK_SERVE`, `DK_ADDR`
- `DK_MODE`, `DK_TARGET`, `DK_PASSWORD`
- `DK_THREADS`, `DK_TIMEOUT`, `DK_RATE`
- `DK_START_PORT`, `DK_END_PORT`, `DK_LATENCY_PORT`
- `DK_FORMAT`, `DK_OUTPUT`, `DK_LOG`, `DK_WORDLIST`
- `DK_VERBOSE`, `DK_SILENT`
- `DK_HEALTH_URL`, `DK_HEALTH_TIMEOUT`

## REST API

- `GET /api/health`
- `GET /api/http?target=example.com`
- `GET /api/plugins/list`
- `GET /api/plugins/results`
- `POST /api/plugins/run`
- `POST /api/scan`

### Run Plugin (example)

```json
{
	"plugin": "http",
	"target": "example.com",
	"timeout_seconds": 3,
	"run_all": false
}
```

## Web UI

- Dashboard: `http://127.0.0.1:8080/`
- Live updates page: `http://127.0.0.1:8080/live`

## Docker

```bash
docker build -f docker/Dockerfile -t defensekit .
docker run --rm -p 8080:8080 defensekit
```

## Docker Compose (Production-style)

```bash
cp .env.example .env
docker compose up -d --build
docker compose ps
docker compose logs -f
```

Stop:

```bash
docker compose down
```

## CI

GitHub Actions workflow: `.github/workflows/ci.yml`

- gofmt check
- go build
- go test

## Smoke Test Script

```powershell
./scripts/smoke.ps1
./scripts/smoke.ps1 -Target scanme.nmap.org -WebPort 8081 -Password "StrongP@ssw0rd123!"
```
