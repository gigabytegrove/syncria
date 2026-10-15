# Syncria v0.8.5 — Administrator sign-in and session reliability

This maintenance release preserves the existing synchronization engine, connections, mappings and data.

- Hide workspace status, live-activity controls, sync data, and dashboard polling until the administrator has signed in.
- Use a dedicated responsive sign-in and initial-setup layout, with properly aligned logo, fields, controls, and spacing.
- Add optional **Remember me**, backed by a signed HttpOnly browser cookie, for 30 days.
- Standard sessions use browser-session cookies with a seven-day maximum on the server.
- Redirect dashboards on session expiry rather than leaving their data visible after an HTTP 401 response.
- Prevent previous authenticated action notices from appearing on the sign-in page.
- Preserve HMAC cookie verification and same-origin POST protection. Signing out clears both session modes.

**Upgrade:** Use the App Updates screen only after the v0.8.5 GitHub release has been confirmed to contain the downloadable assets and checksums. Docker users should use `ghcr.io/gigabytegrove/syncria:v0.8.5` only after that image has been published. Existing /data configuration stays intact.
