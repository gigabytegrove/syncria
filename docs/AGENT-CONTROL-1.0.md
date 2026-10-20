# Syncria 1.0.0 — Device identity API (implementation)

**This feature does not enable filesystem access or device-to-device sync.** It is an authenticated, durable identity subsystem that is safe to integrate ahead of the transfer engine.

## Server integration

Place `agentcontrol/`, `agent_control_routes.go` in the root of `github.com/gigabytegrove/syncria`. Insert the following immediately after `mux := http.NewServeMux()` in `main.go`:

```go
if err := registerAgentAPIs(mux, a, dir); err != nil {
    log.Fatal(err)
}
```

Identity state is stored as `agents.json` in Syncria's existing private data directory. An invalid state file prevents startup, without resetting existing cloud/NAS configuration. The administrator must be signed in to issue and approve enrollment, see devices and audit events, or revoke a device.

## API

- `POST /api/v1/agents/invitations` (admin cookie + same-origin POST): JSON `{ "name": "Office PC", "ttl_seconds": 600 }`. Returns invitation, **single-display** pairing token.
- `POST /api/v1/agents/invitations/{id}/approve` (admin cookie + same-origin POST): explicit approval.
- `POST /api/v1/agents/enroll` (TLS): JSON `{ "token": "...", "platform": "windows", "arch": "amd64", "version": "1.0.0" }`. Returns an ID and **single-display** device secret.
- `POST /api/v1/agents/heartbeat` (TLS): `X-Syncria-Agent-ID: <id>` and `Authorization: Bearer <device-secret>`; JSON `{ "platform":"windows", "arch":"amd64", "version":"1.0.0" }`.
- `GET /api/v1/agents/devices`, `GET /api/v1/agents/invitations`, `GET /api/v1/agents/audit` require admin authentication.
- `POST /api/v1/agents/devices/{id}/revoke` (admin cookie + same-origin POST): permanent device credential revocation.

Credential endpoints require HTTPS; the handler permits loopback only for local testing. A TLS-terminating reverse proxy can set `SYNCRIA_TRUSTED_HTTPS_PROXY=1` provided it **overwrites** `X-Forwarded-Proto` and blocks direct access to Syncria's server listener. Do not set that variable otherwise. No remotely supplied agent path can be browsed or read using this API.

## Verification

Run `go test -race ./agentcontrol` on Go 1.25+ and `go test ./...` in the full repo. The stand-alone agentcontrol tests were run locally using Go 1.23.2 (standard-library-only). Full-application compile and runtime tests remain required before merging.
