# Tailwind development
Syncria ships a precompiled, local Tailwind stylesheet embedded in the Go binary. No internet connection, browser CDN, npm process, or separate CSS service is required at runtime.

Install the UI development dependency with `npm install`, compile with `npm run build:css`, then use `go test ./...` and `go build .`. The Dockerfile performs the stylesheet build in its own stage before compiling Go.

Edit `assets/syncria.tailwind.css` rather than editing the generated `assets/syncria.css` directly. UI hooks and form action names in `ui.go` are part of the running application. Do not remove them during layout work.

Design approach: neutral light surfaces, modest spacing, standard controls, responsive navigation, and readable statuses. Sync safety features, the four-step wizard, recovery controls, NAS connections, account linking, file picker, updater and activity reporting remain present.
