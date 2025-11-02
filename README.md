# Search Service (Go + PostgreSQL)

Minimal, case-aligned search aggregator:
- 2 providers (JSON + XML)
- Normalize → Score → Sort → Paginate
- Provider-level rate limit + timeout
- In-memory TTL cache (120s)
- UPSERT to PostgreSQL
- Simple web dashboard

## Quick start
```bash
docker-compose up --build
```
- Dashboard: http://localhost:8080/dashboard/index.html
- API: curl 'http://localhost:8080/search?query=istanbul&type=video&sort=popularity&page=1&size=10'

### ENV
- DB_DSN (compose default provided)
- PROVIDER1_URL / PROVIDER2_URL (defaults use local mock files under /app/web/mock/)
- CACHE_TTL_SECONDS (default 120)
- RATE_LIMIT_PER_MINUTE (default 60)
