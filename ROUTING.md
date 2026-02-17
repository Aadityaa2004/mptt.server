# Routing Architecture

This document describes how nginx routes requests to the frontend (Next.js) vs the backend (Go API). **When adding new API endpoints, you must update both the backend and nginx config.**

## Quick Reference

| Path | Proxied to | Notes |
|------|------------|-------|
| `/api/geocode` | `mqt-frontend:3000` | Next.js API route (Nominatim proxy) |
| `/api/*` | `api-service:9002` | Auth, users, locations (Go) |
| `/pis`, `/pis/*` | `api-service:9002` | Pi CRUD (Go) |
| `/stats`, `/stats/*` | `api-service:9002` | Stats endpoints (Go) |
| `/readings`, `/readings/*` | `api-service:9002` | Readings (Go) |
| `/health`, `/health/*` | `api-service:9002` | Health checks (Go) |
| `/mqtt` | `mosquitto:9001` | MQTT WebSocket |
| `/` (all other) | `mqt-frontend:3000` | Next.js frontend |

## Adding a New Endpoint

### Backend (Go API) routes

1. Add the route in the appropriate controller under `src/production/MQT.ApiService/controllers/`.
2. **If the path has no `/api` prefix** (e.g. `/new-resource`):
   - Add a new `location /new-resource` block in both:
     - `nginx/default.conf` (HTTPS server block)
     - `nginx/default-http-only.conf`
   - Proxies to `api-service:9002`.
3. **If the path is under `/api`** (e.g. `/api/new-resource`):
   - If it's a **Next.js API route** (proxy to external service, server-side logic):
     - Add `location /api/new-resource` **before** `location /api` in nginx.
     - Proxy to `mqt-frontend:3000`.
   - If it's a **Go backend route**:
     - No nginx change needed; `location /api` already catches it.

### Frontend API calls

- `apiClient.apiFetch()` uses `API_BASE_URL` from `constants/api.ts`.
- Paths are appended as-is: `API_BASE_URL + "/pis"` → `https://domain.com/pis`.
- Ensure the path matches what nginx routes to the API (see table above).

## Files to Update

| Change | Update these files |
|--------|--------------------|
| New Go route without `/api` | `nginx/default.conf`, `nginx/default-http-only.conf` |
| New Next.js API route under `/api` | `nginx/default.conf`, `nginx/default-http-only.conf` (add before `/api`) |
| New Go route under `/api` | None (already proxied) |

## Nginx Location Order

Nginx matches locations by **longest prefix first**. Order matters:

1. `/api/geocode` must appear **before** `/api` so geocode goes to Next.js.
2. Specific paths (`/pis`, `/stats`, `/readings`) must appear before `/`.
