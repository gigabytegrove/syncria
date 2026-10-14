# Syncria v0.5.3 — Cleaner updates and a more finished interface

This update fixes two things that made Syncria feel unfinished: completed updates staying on screen forever, and dashboard pages that did not fit together visually.

## Updates now get out of your way

While an update is running, you can still see the **real download progress, file checks, installation steps, and restart status**.

When the new Syncria version starts successfully, the screen confirms that it worked, then returns to the usual App Updates page. The completed update no longer stays in your activity panel, and its finished status is removed from saved update state.

If an update fails, Syncria **keeps the error visible** so you can see what happened and try to fix it. A successful update and a failed update are treated differently.

## A more polished dashboard

- A consistent layout, spacing, colors, buttons, forms, and navigation on every page.
- Clearer Home screen sections for your connections, sync activity, and folders.
- The getting-started guide stays prominent for new users, but takes less space after you've set up a sync.
- Each sync has its own easy-to-scan card with its controls grouped together. Progress is shown while a sync is running, not as an empty bar all the time.
- Better organized cloud accounts, NAS boxes, and activity information.
- The Updates page stays clean when nothing is running. Detailed progress and recorded events appear only when needed.
- More consistent small-screen layout and keyboard focus indicators.

The plain-language setup and help from v0.5.2 stay in place. This is a design and updater cleanup, **not a change to the sync engine**.

## What stays the same?

Your saved accounts, NAS connections, files, sync settings, conflict choices, 10-minute Undo options, and update process remain the same. Syncria does not save copies of your files on its server.

## Install the update

Open **App Updates → Check again → Download & install update** in your existing Syncria dashboard. There is no need to build Syncria yourself or replace the container.

**One-time note:** The page that installs v0.5.3 is still running the older v0.5.2 interface in your browser. After the update, refresh the page once to see the new design and automatic success cleanup. Future updates will use the new behavior.

[Getting Started](GETTING-STARTED.md) explains everyday setup. [Technical Reference](TECHNICAL.md) is available for advanced server and connection details.
