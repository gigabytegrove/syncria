# Syncria CSS build

Tailwind CSS is compiled during development and must be bundled locally for offline operation. Install Node development dependencies with `npm ci`, then run `npm run build:css`. Never use a browser CDN or require Node on a user's NAS.

The source stylesheet is `assets/syncria.tailwind.css`; `assets/syncria.css` is the compiled output. The embedded Go UI must use the compiled stylesheet. The existing template IDs, classes used by JavaScript, and all safety-related forms must be retained.
