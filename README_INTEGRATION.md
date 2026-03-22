# UniMatch Integration Guide

This project now uses a unified Docker setup to manage the Go Backend, Python AI Service, Neo4j Graph Database, and Frontend.

## Infrastructure Map

- **Frontend**: Nginx on `:80` (Host `:5173`). Proxies `/api/` to Backend.
- **Backend (Go)**: API on `:8894`. Communicates with PostgreSQL and AI Service.
- **AI Service (Python)**: API on `:8895`. Handles workers, TinyFish agents, and Neo4j.
- **Neo4j**: Graph DB on `:7474` (HTTP) and `:7687` (Bolt).
- **Claude Proxy**: Internal (Docker network only) on port `:4000`.

## Step-by-Step Instructions

### 1. Unified Startup (Recommended)
From the root directory, run:
```bash
docker-compose up -d --build
```
This will start all components in the correct order. The backend will wait for the AI service, and the AI service will wait for Neo4j.

### 2. Synchronized Deletion
When you delete a university in the **University KB** view:
1. The **Backend** receives the request.
2. It calls `DELETE /jobs/university/{id}` on the **AI Service**.
3. The AI service deletes the university and all related orphan nodes/edges from **Neo4j**.
4. The Backend deletes the record from **PostgreSQL**.

### 3. Environment Synchronization
- All containers are on the same bridge network.
- Use `ai-service:8895`, `db:5432`, and `neo4j:7687` for internal container-to-container communication.
- External tools (like MCP on your host) are reachable via `host.docker.internal`.

## Configuration Files
- `docker-compose.yml`: Main orchestrator.
- `ai_service/.env`: AI service secrets and Neo4j uri.
- `backend/.env`: Backend db and ai service urls.
- `frontend/.env`: Frontend api path.

## Development Tools

### Lite Data Seeding
If you want to quickly populate the database with a small set of test data (3 universities, 1 admin, 1 teacher):
```bash
cd backend
go run scripts/lite_seed.go
```
This is useful for local UI testing without waiting for a full crawl or long analysis.
