# Syncria v0.6.5 — an updater that makes sense

The App Updates screen has been rebuilt around the information you actually need.

## A clear update status

- **Update available:** a prominent status message, the new version number, and a direct **Download and install** button.
- **Up to date:** a clear confirmation and the installed version. No unexplained disabled install controls.
- **Release pending or check failed:** specific status and guidance instead of a misleading current-version message.
- **New** indicator in the navigation when a verified newer release is available. Syncria checks at startup and refreshes periodically.
- A direct link to the release notes so you can review what changes before installing.

## A simpler layout

- Removed the confusing permanent right column, the unnecessary "How it works" panel, and decorative step descriptions.
- Put installation progress, verification, and restart information directly below the main status panel **only when needed**.
- Kept the real update log available during installation.
- Moved manual file installation and **Restore previous version** into collapsed **Advanced options** at the bottom.
- Improved responsive behavior without adding marketing-style effects.

## Unchanged

Updates still download and verify the correct platform binary, retain rollback support, and run entirely inside Syncria. No external updater container, Docker platform dependency, or internet-hosted UI assets are introduced.

Completed-update cleanup and failed-update reporting continue to follow the existing safeguards.
