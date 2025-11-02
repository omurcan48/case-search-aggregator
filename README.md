# Case Search Aggregator (Go + PostgreSQL)

3 sağlayıcıdan (JSON + XML) gelen içerikleri **normalize → skorla → sırala → sayfala** akışıyla birleştiren basit arama servisi.
Not: 3 sağlayıcıyı arttırabilirsiniz ya da azaltabilirsiniz.
Tek komutla Docker Compose ile ayağa kalkar; Postgres’e **UPSERT** yazar, ön yüzde basit bir dashboard sunar.

- Go + Chi HTTP
- PostgreSQL (pgx driver)
- In-memory TTL cache
- Sağlayıcı bazlı rate-limit + timeout
- Dockerfile + docker-compose
- Basit web dashboard (`/web/dashboard/index.html`)

---

## ⚡ 3 Dakikada Hızlı İnceleme

```bash
docker-compose up --build
# Health:     http://localhost:8080/health
# Dashboard:  http://localhost:8080/dashboard/index.html
# Örnek API:  http://localhost:8080/search?query=istanbul&type=video&sort=popularity&page=1&size=10

