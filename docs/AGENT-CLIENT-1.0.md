# Syncria 1.0.0 development — Device companion agent

**Status: functional enrollment, persistent identity, and heartbeat; synchronization is not yet enabled.** This is an explicitly opt-in Go executable. It cannot browse, read, upload, download, delete or transfer files, and it does not self-update yet.

## Build

From the Syncria repository root using Go 1.25+:

```bash
go test -race ./agentcontrol ./agentclient ./cmd/syncria-agent
go build -trimpath -o syncria-agent ./cmd/syncria-agent
```

On Windows PowerShell, run `go build -trimpath -o syncria-agent.exe ./cmd/syncria-agent`.

## Enroll a device

1. Sign into the Syncria administrator UI. Use the authenticated invitations API to create a token with an expiration. Copy the token; it is displayed only once.
2. Approve its invitation ID through the authenticated approval API.
3. Run `syncria-agent enroll -server https://syncria.example`. Paste the token into standard input when prompted. Agent registration returns an identity that is saved under the platform's user configuration directory in `syncria-agent/identity.json`.
4. Run `syncria-agent run` as the enrolled operating-system user to start heartbeat reporting. It sends immediately and then approximately once per minute when healthy. To check one heartbeat and exit, run `syncria-agent run -once`.
5. Use `syncria-agent status` to view device name, ID, server and enrollment date without exposing credentials.

The agent intentionally does **not** automatically run as a service yet. A service installer, signed update channel, transfer operations and configured folder permissions require their own implementation and acceptance. Use HTTPS for the server; plain HTTP is permitted only for loopback development. TLS certificates must be trusted by the operating system. Redirects are not followed for any enrollment or credential request.

## API and lifecycle

See `docs/AGENT-CONTROL-1.0.md`. The server grants a distinct secret for each approved device; token codes expire and cannot be redeemed again. On each heartbeat, the server checks the credential and revocation state. The local secret is never printed by `status` or successful heartbeat output. If administrator revokes a device, an agent exits after its next rejected heartbeat. Existing NAS/cloud jobs and the default single-server installation are unaffected.

## Security boundaries and deployment caveats

- Treat the pairing token as secret until redeemed. Do not include it in shell command-line arguments or publish it to logs.
- The identity file is owner-only on Unix. Running the agent as a different user will not automatically find the credentials.
- For TLS-terminating reverse proxies, only set `SYNCRIA_TRUSTED_HTTPS_PROXY=1` when the proxy overwrites incoming `X-Forwarded-Proto` and Syncria's direct HTTP listener is inaccessible to untrusted clients.
- This is a **device identity** capability, not a file synchronization agent. No file access is granted by enrollment. Future authorized roots must be approved independently and transfers must use race-safe path APIs.
- A lost credential requires administrator intervention. There is intentionally no unauthenticated password recovery or secret retrieval endpoint.
- Protect and back up the controller's `agents.json` alongside its existing `/data`, using restrictive filesystem permissions. Corrupt or insecure agent identity state stops startup rather than silently discarding device authorization data.
