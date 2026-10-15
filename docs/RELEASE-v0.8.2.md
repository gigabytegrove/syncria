# Syncria v0.8.2 — A cohesive synchronization workspace

This release redesigns the user experience of existing synchronization features. The sync engine, timestamp-first comparison rules, deletion semantics, and on-change scheduler are unchanged from v0.8.1.

## My Syncs: find what matters
- Compact, scannable connection cards with meaningful status, direction, scheduling, latest result, and file totals. No more long rows of competing primary buttons.
- Search and filter syncs by name or state (all, running, needing attention, automatic enabled, paused).
- **Open sync** launches the real detail page. **Sync now** is prominent; run cancellation is visible only while running.
- Less common actions live together under a contextual **More** menu: read-only preview, edit configuration, pause/enable checks, and remove connection.
- Existing unresolved conflict controls and recoverable deletion actions stay available on the cards so critical safety choices are not hidden in a menu.

## One place to understand a sync
Open a sync and use the four dedicated tabs:

1. **Overview:** Live status, recent quick-check state, successful checks, total transfers, latest recorded run, file direction, and deletion behavior.
2. **Files & preview:** On-demand read-only comparison showing actual proposed copies, replacements, skips and deletion/conflicts with reasons. The panel is attached to the selected sync and does not run automatically upon opening a page.
3. **Run history:** Select any run to see its outcome, duration, bytes, and recorded per-file operations. Search complete stored file paths and filter by action server-side, not just among the current 100 visible records; paginate older runs and longer file lists.
4. **Settings:** Current transfer direction, schedule, deletion sharing, and recovery behavior explained in plain terms, with Edit and Preview actions.

Sync-detail URLs are bookmarkable as `#sync-detail/<id>/<overview|files|history|settings>`, with safe restoration after a browser refresh.

## Consistent polish
- Shared flat, responsive visual language, deliberate spacing, scoped primary and secondary actions, accessible tab and filter labels, quiet empty states, readable long paths, and keyboard-dismissable More menus.
- Desktop and mobile Chromium screenshot checks run alongside Go, race, static-analysis, and JavaScript parser/DOM checks.
- Historical failure totals are not misleadingly labeled active attention; live alerts reflect present state.

## Important safety notes
- No changes to sync direction, deletion policy, existing index, checkpoint format, account credentials, NAS shares, or scheduled/background execution.
- **Preview never performs transfers** and only runs when you explicitly request it.
- Dismissing a previous failure remains separate from erasing historical run counts.
- Run history is complete for operations recorded by v0.8.1 and later. Earlier versions' lost per-file histories cannot be reconstructed.
- Back up Syncria's persistent `/data` directory before installing any updater. v0.8.x retains its existing database and backup rules.

## Suggested smoke test
1. With a relationship paused, update through App Updates and confirm v0.8.2.
2. Open My Syncs, filter by name and state, then click a connection.
3. Visit each tab and review deletion settings. Choose Preview and confirm it does not modify NAS/Drive files.
4. Select the previous run; search for a known transferred filename, including one beyond the first 100 events.
5. Test Sync now, Stop, Pause and Edit; confirm correct state and that historical records persist after refresh.
6. Check desktop and mobile sizes. Report any unexpected overlap, missing controls or incorrect context menu behavior.
