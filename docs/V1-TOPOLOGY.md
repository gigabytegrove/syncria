# Syncria 1.0 – live configuration topology

The Sync Map screen is read-only and is part of the existing Syncria dashboard. Data is served to **authenticated administrators only** from `GET /api/v1/sync/topology`. No independent topology database, file explorer, cloud relay or endpoint agent is introduced.

- Device nodes come from the persistent approval registry. "Recently seen" means a heartbeat was accepted within two minutes; it **does not** assert a transfer is running or the device is reachable now. Revoked devices remain labeled revoked.
- NAS nodes come from saved NAS shares; cloud nodes come from configured Google/Dropbox accounts. No credential, token, absolute local folder path or cloud refresh token is serialized into the graph.
- Edges come only from actual configured `agent_bridge`, `nas_bridge`, and `bridge` relationships. Disabled jobs show paused; jobs with errors show attention. Missing endpoints are warnings, never invented devices or connections.
- Older local-storage sync relationships lack graph endpoint metadata and are currently excluded with a warning. The graph is **not yet complete parity** with every historical job provider.
- Node names and relationship labels are set using DOM `textContent`, not `innerHTML`.

## Verification

`go test -race -count=3 ./synctopology` and `go vet ./synctopology` passed in the isolated Go 1.23.2 workspace. The new dashboard JavaScript passed `node --check`. Full Go 1.25 controller tests and browser/device regression tests remain required before marking platform-wide alpha complete.
