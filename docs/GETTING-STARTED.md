# Getting started with Syncria

Welcome! Syncria keeps copies of your files in the places you choose. You can use it with Google Drive, Dropbox, or one or more NAS boxes.

A **NAS** is simply a storage box on your home or work network. It usually holds folders just like a computer does.

## The three things to do first

### 1. Tell Syncria where your files are

If you use Google Drive or Dropbox, open **Cloud Accounts** and select **Add a cloud account**. Choose the service and follow the sign-in steps.

If the sign-in page says it needs extra setup, open **Settings**. Someone who manages your Syncria installation may need to set up your cloud application's sign-in details first. That is a one-time administrator task, not something you need to learn for everyday use.

If your files are only on NAS boxes, you do not need to connect a cloud account.

### 2. Add your storage box

Open **My NAS Boxes** and click **Add a NAS box**. Syncria will ask for:

- **A name you like** — for example, “Family NAS.”
- **The NAS address** — often a set of numbers like `192.168.1.20`.
- **The shared folder name** — such as “Pictures” or “Family.”
- **The username and password** for that NAS.

You can find these details in the NAS settings or ask whoever set up the storage box. Click **Test and save my NAS**. Syncria should tell you whether it can open the shared folder.

### 3. Make your first sync

Open **My Syncs** and click **Set up a sync**. Answer four simple questions:

**Where are my files now?** Choose the cloud account or first NAS box that already has the files.

**Where do I want copies?** Choose the NAS folder or folders where you want them.

**How should it work?** Pick when to check. The usual settings are already selected. For two NAS boxes, you can also choose whether copies move one way or both ways.

**Does everything look right?** Read the summary and save.

Your new sync starts **turned off**. Nothing is copied until you click **Turn on sync**. You can use **Sync now** to ask Syncria to check right away.

## Examples

**“I want my Google Drive pictures on my NAS.”** Choose “My cloud files and a NAS box,” select Google Drive, then select the NAS folder where copies belong.

**“I have two NAS boxes and want them to match.”** Choose “Two NAS boxes.” Select which NAS has the files first, choose the other NAS, then decide whether updates should go only one way or both ways.

**“I want to keep two different copies.”** In a two-way NAS sync, choose “Save both copies with different names” under the option for different versions of the same file.

**“I want Syncria to ask before choosing a copy.”** Choose “Ask me which copy to keep (safest).” When there is a conflict, open **My Syncs** to make the choice.

## If I delete a file, what happens?

By default, when Syncria notices that a file was deleted in one place, it will remove the matching copy in the other connected place too.

Syncria gives you **10 minutes to bring it back** after it notices the deletion. Open **My Syncs** and use **Bring back file** if the deleted item appears there. The recovery copy stays on your NAS or cloud service, not on the Syncria computer.

During setup, you can also turn off deleting matching copies or turn off the Undo option for that sync. If you are not sure, leave the Undo option on and keep a separate backup of important files.

## When two files are different

If you edit the same file on two NAS boxes before they have copied changes, Syncria needs to know which version to use. Here are the five available choices:

| What Syncria shows | What it means |
|---|---|
| Ask me which copy to keep | Syncria stops on that file so you can choose |
| Use the copy changed most recently | The latest modified time decides |
| Always choose the first NAS copy | The first NAS is the winner |
| Always choose the other NAS copy | The second NAS is the winner |
| Save both copies with different names | Both files are kept |

These choices apply to NAS-to-NAS syncs that work both ways.

## If Syncria says something went wrong

**“Could not connect.”** Check that the NAS is on, its address is correct, and the username/password are right.

**“Could not open folder.”** Make sure the folder exists and the account you used can read it.

**“Two copies are different.”** Open **My Syncs**, read the filenames, and choose what Syncria should keep.

**“Update failed.”** Open **App Updates**. Read the last recorded step and the error. You should see exactly how far it got.

If a message includes technical details, you can share those with the person who manages your NAS or Syncria server.

## How do I update it?

Use **App Updates → Check again → Download & install update**. The screen shows the actual download and file-checking progress, then waits for Syncria to restart and report the new version.

You don't need to stop Docker, run a compiler, or install another updater.

## A few words you may see

| Word | Meaning |
|---|---|
| Cloud | Files online, such as Google Drive or Dropbox |
| NAS | A file storage box on your network |
| Sync | Keeping matching copies of files |
| One way | Copy changes from the first NAS to the other NAS |
| Both ways | Changes can go in either direction |
| Conflict | Two copies of a file are different |
| Undo | Bring back an accidentally deleted file within the allowed time |

For administrator and developer details, see [Technical reference](TECHNICAL.md).