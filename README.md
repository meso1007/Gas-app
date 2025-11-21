# Backend Documentation

## Overview
The **backend** of the Gas‑Insight application is a Go server that:
- Fetches news articles from external APIs.
- Analyzes the content using the Gemini AI model.
- Stores the results (summary, sentiment, etc.) in a SQLite database.
- Exposes a JSON REST API for a future frontend to consume.

## Project Structure
```
backend/
├─ cmd/                 # Entry points (e.g., local server, CLI tools)
│   └─ local/           # `main.go` for local development
├─ internal/            # Core packages
│   ├─ fetch/           # News fetching logic
│   ├─ detect/          # Gemini analysis logic
│   └─ db/              # SQLite DB helpers
├─ .env                 # Environment configuration (example file provided)
├─ .gitignore           # Backend‑specific ignore rules
├─ go.mod, go.sum       # Module definition
└─ README.md            # **This file** – detailed backend docs
```

## Getting Started
### Prerequisites
- Go 1.22+ installed (`go version` should show 1.22 or later).
- (Optional) `sqlite3` command‑line tool for inspecting the DB.

### Setup
1. **Clone the repository** (if you haven’t already):
   ```bash
   git clone git@github.com:meso1007/Gas-app.git
   cd Gas-app/backend
   ```
2. **Install Go module dependencies**:
   ```bash
   go mod tidy
   ```
3. **Create an `.env` file** (copy from `.env.example` if present) and set the required variables:
   ```dotenv
   # .env
   NEWSAPI_KEY=your_newsapi_key_here
   GEMINI_API_KEY=your_gemini_api_key_here
   DATABASE_PATH=data/gasinsight.db   # default location
   PORT=8080
   ```
4. **Run the server locally**:
   ```bash
   go run ./cmd/local
   ```
   The server will start on `http://localhost:8080`.

## API Endpoints
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Simple health check – returns `{"status":"ok"}` |
| `GET` | `/news` | Returns the latest fetched news items stored in the DB |
| `POST` | `/fetch` | Triggers a manual fetch from the external news source (useful for testing) |
| `POST` | `/analyze` | Accepts a JSON payload `{ "url": "https://..." }` and returns Gemini analysis results |

All responses are JSON and include a `code` field for HTTP status and a `data` field for the payload.

## Core Packages
- **`internal/fetch`** – Implements `FetchNews()` which calls the NewsAPI, parses the response, and stores raw articles in the DB.
- **`internal/detect`** – Contains `GeminiAnalyzer` that sends article text to the Gemini API and parses the summary/sentiment.
- **`internal/db`** – SQLite helper functions (`OpenDB`, `InsertArticle`, `GetArticles`, etc.).

## Testing
Unit tests are located alongside each package (e.g., `fetch/fetch_test.go`). Run them with:
```bash
go test ./... -v
```
Integration tests that hit external APIs are skipped by default; set the `INTEGRATION=true` env var to enable them.

## Deployment
The backend can be containerised with Docker. A minimal `Dockerfile` example:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gasinsight ./cmd/local

FROM alpine:latest
WORKDIR /app
COPY --from=builder /gasinsight .
COPY .env.example .env
EXPOSE 8080
CMD ["./gasinsight"]
```
Build and push the image, then run it on any container platform (Docker, Cloud Run, etc.).

## Contributing
1. Fork the repo and create a feature branch.
2. Follow the existing code style (gofmt, go vet, staticcheck).
3. Write tests for new functionality.
4. Open a Pull Request with a clear description.

---
*Last updated: 2025‑11‑21*
