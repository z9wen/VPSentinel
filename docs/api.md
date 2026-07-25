# API and token model

All JSON endpoints reject oversized request bodies. Admin routes use
`Authorization: Bearer <access JWT>`. Agent routes use a node-specific bearer
token. Error responses have a stable `error.code` and human-readable message.

## Admin authentication

1. `POST /api/v1/auth/login` with `{"email":"...","password":"..."}`.
2. Keep the returned access JWT in memory and send it on admin API requests.
3. When it expires, call `POST /api/v1/auth/refresh` with credentials enabled.
   The HttpOnly refresh cookie is rotated and a new access JWT is returned.
4. `POST /api/v1/auth/logout` revokes the current refresh session.

The browser never reads the refresh token. This reduces exposure to XSS and
makes server-side session revocation possible without putting every access JWT
lookup on the database hot path.

## Main endpoints

| Method | Path | Authentication | Purpose |
|---|---|---|---|
| POST | `/api/v1/agent/metrics` | Node token | Store an idempotent metric batch |
| POST | `/api/v1/agent/heartbeat` | Node token | Update last-seen without a sample |
| POST | `/api/v1/auth/login` | Public | Create an access/refresh pair |
| POST | `/api/v1/auth/refresh` | Refresh cookie | Rotate the refresh session |
| POST | `/api/v1/auth/logout` | Refresh cookie | Revoke the refresh session |
| GET/POST | `/api/v1/nodes` | Access JWT | List or register nodes |
| GET/DELETE | `/api/v1/nodes/:id` | Access JWT | Inspect or delete a node |
| POST | `/api/v1/nodes/:id/tokens` | Access JWT | Generate a token shown once |
| GET | `/api/v1/nodes/:id/tokens` | Access JWT | List token metadata, never plaintext |
| DELETE | `/api/v1/nodes/:id/tokens/:tokenId` | Access JWT | Revoke a node token |
| GET | `/api/v1/nodes/:id/metrics/latest` | Access JWT | Fetch latest host metrics |
| GET | `/api/v1/alerts?active=true` | Access JWT | Fetch active or historical alerts |
| GET | `/api/v1/alerts/:id` | Access JWT | Fetch one alert and raw webhook |
| POST | `/api/v1/integrations/grafana/webhook` | Webhook secret | Store firing/resolved events |
| GET | `/healthz`, `/readyz`, `/version` | Public | Liveness, readiness, build info |

Grafana webhook requests send the secret as a Bearer credential; the API also
accepts `X-VPSentinel-Webhook-Secret` for manual integration tests. A
fingerprint, state, and start timestamp form
the idempotency key so a retried notification does not duplicate an incident.
