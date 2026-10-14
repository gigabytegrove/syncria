# Syncria

**Keep your files together, wherever you save them.**

**Interface (v0.6.3):** Syncria uses a quieter, locally bundled Tailwind CSS design with straightforward navigation, forms, progress, and storage management. No browser CDN or external UI service is needed.

Syncria helps you keep folders matched between Google Drive, Dropbox, and storage boxes on your home or work network. Those storage boxes are called **NAS devices**.

For example, you can have family photos in Google Drive and keep a matching copy on your NAS. Or you can keep two NAS boxes matched without using the cloud at all.

Syncria does not become another place that stores your files. Your files stay in your cloud account or on your NAS.

## Getting started

You don't need to know networking terms to use Syncria's everyday screens. Once Syncria is installed, open it in your browser and follow these steps:

1. **Cloud Accounts:** If your files are in Google Drive or Dropbox, connect the account you use. If you're only copying between NAS boxes, skip this step.
2. **My NAS Boxes:** Add the storage box you want to use. You'll need its network address, shared folder name, username and password.
3. **My Syncs → Set up a sync:** Choose **where your files are now**, **where copies should go**, and **how often Syncria should check**. Read the final summary and save it.
4. **Turn on sync:** New syncs start **off**, so they will not copy anything automatically. You can turn a sync on or choose **Sync now** when you're ready.

For examples and explanations of common choices, read [Getting Started](docs/GETTING-STARTED.md).

## What can Syncria do?

- **Google Drive ↔ NAS:** Keep cloud files and a NAS folder matched.
- **Dropbox ↔ NAS:** Do the same with a Dropbox account.
- **NAS ↔ NAS:** Keep folders on separate NAS devices together, without needing any cloud account.
- **More than one account or NAS:** Set up separate syncs for different family members or devices.
- **One way or both ways:** For NAS-to-NAS syncs, decide whether changes travel in one direction or both.
- **Handle two different copies:** When the same file was changed on both NAS boxes, you can tell Syncria to ask you, use the newer copy, choose one NAS, or keep both copies.
- **Bring back an accidental deletion:** By default, Syncria gives you a **10-minute Undo window** after it finds a deletion, when the storage service supports recovery. You can change the deletion settings for each sync.

### Important things to know

**A sync is not a separate backup.** If you delete a file and deletion sharing is on, Syncria can remove its matching copies too. Keep an independent backup of files you cannot replace.

**The 10-minute clock starts when Syncria detects the deletion**, which may be later than when you pressed Delete. Recovery depends on the storage service being available. Syncria does not secretly save your files on its server.

**The app will not automatically overwrite a file it cannot safely handle.** Some cloud-native items, shared drives, shortcuts, and complicated moves need special care. If a file cannot be handled safely, you should see a problem instead of a false success.

## Updating Syncria

You normally update from **inside Syncria**:

1. Open **App Updates**.
2. Click **Check again**.
3. Click **Download & install update**.
4. Watch the real download and checking progress. Syncria will restart and confirm the new version.

Your saved accounts, NAS boxes, and sync choices remain in place. You do **not** need to build an app, download a replacement image, or add an extra updater container each time.

## Installing Syncria for the first time

Syncria is self-hosted. An adult or person who manages your home server will need to install it once. After that, most actions happen in the browser.

The repository includes a [Docker Compose configuration](compose.yaml) and [Dockerfile](Dockerfile); standalone binaries are attached to [GitHub Releases](https://github.com/gigabytegrove/syncria/releases). The Compose image depends on GitHub Container Registry access, which may be restricted while package-publishing permissions are being resolved. If the image cannot be pulled, do not assume the app is broken.

[Technical installation instructions and older release details](docs/TECHNICAL.md) are available for whoever manages the server. The **in-app updater** is still the main way to update an existing installation.

## Need help?

[Getting Started](docs/GETTING-STARTED.md) explains the everyday screens, common choices, and what to check when a connection fails.

[Latest changes](docs/RELEASE-v0.6.0.md) explains what's different in the newest version without developer jargon.

[Technical reference](docs/TECHNICAL.md) keeps advanced deployment, API, security, and release history details for administrators.

**Made by [Gigabyte Grove](https://github.com/gigabytegrove).**