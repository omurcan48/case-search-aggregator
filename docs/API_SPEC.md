# API Spec

## GET /search

Query params:
- `query` (required)
- `type` = `video` | `text` (optional)
- `sort` = `popularity` | `relevance` (optional; default popularity)
- `page` >= 1 (default 1)
- `size` in [1..100] (default 20)
