# ai-service-go

Simple AI orchestration service for UniMatch.

## Features

- Accepts async analyze, crawl, and report jobs from `unimatch-be`
- Uses `Exa` first as the flexible multi-query search layer
- Uses `TinyFish` next to drill into detail pages and extract the exact fields needed
- Retries search flexibly with up to 5 Exa attempts per content item
- Falls back to `OpenAI` JSON synthesis to fill missing fields when evidence is still incomplete
- Keeps deterministic heuristics as the final safety net so the full backend flow still works locally
- Sends results back through the existing backend callback contract

## Endpoints

- `POST /jobs/analyze`
- `POST /jobs/crawl`
- `POST /jobs/report`
- `GET /jobs/:job_id` (local/test environments only)
- `GET /health`
- `GET /swagger/index.html`

## Run

```powershell
go mod tidy
go run main.go
```

The service listens on port `9000` by default, matching the backend's `AI_SERVICE_URL`.

## Local Debug And Retry Controls

- `OPENAI_RETRY_ATTEMPTS=5`
- `CALLBACK_RETRY_COUNT=3`
- `CALLBACK_RETRY_DELAY_MS=300`
- `MAX_SEARCH_ATTEMPTS=5`
- `MAX_DETAIL_FETCHES=3`

`GET /jobs/:job_id` is available outside production so local verification can inspect search attempts, TinyFish fetches, OpenAI fill usage, and callback delivery state.
