# Syncria v0.6.1

This release replaces the previous flashy interface with a more practical design. It is meant to look and feel like a file-sync application rather than a marketing dashboard.

## What changed
- Light, restrained navigation and consistent page layouts.
- Clearer Overview, sync management, NAS storage, cloud account, activity, settings and update screens.
- More consistent forms, status messages, setup steps and dialogs.
- Responsive layout for smaller screens without promotional animations or illustrations.
- Tailwind CSS is compiled when Syncria is built, then packaged inside the application. The interface doesn't need an external CSS service or internet access.
- The existing sync engine, schedules, conflict decisions, deletion recovery, cloud connections and built-in updater are unchanged.

## Updating
Install this release through Syncria's built-in updater. Your existing accounts, storage connections and sync definitions remain in place. The optional container image uses the same version.

## Testing
The release workflow compiles Tailwind locally, runs Go tests and race checks, and cross-compiles updater binaries before publishing release assets.
