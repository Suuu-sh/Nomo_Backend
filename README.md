# Nomo Backend

Go API for Nomo. It proxies authenticated requests to Supabase/PostgREST using the caller's Supabase JWT so RLS remains enforced by Supabase.

## Local run

```sh
cp .env.example .env
# set SUPABASE_ANON_KEY from /Users/yota/Projects/Secrets/Nomo/supabase_dev-nomo.md
export $(grep -v '^#' .env | xargs)
go run ./cmd/api
```

Health check:

```sh
curl http://localhost:8080/healthz
```

Authenticated requests must include:

- `Authorization: Bearer <supabase access token>`
- `X-Nomo-User-ID: <auth.users.id>`

## Endpoints

- `GET /healthz`
- `GET /v1/me/profile`
- `PATCH /v1/me/profile`
- `GET /v1/friends`
- `GET /v1/drink-logs`
- `POST /v1/drink-logs`
- `GET /v1/daily-status?date=YYYY-MM-DD`
- `PUT /v1/daily-status`
