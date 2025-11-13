# Onyx


Onyx is a simple, friendly command-line tool for managing your daily notes in an
[Obsidian](https://obsidian.md/) vault. It helps you quickly add and review notes for any day,
right from your terminal.

## ✨ Features

- **Add notes** to your daily Obsidian journal with a single command
- **List notes** for any date, beautifully rendered in your terminal
- **Manage a day planner**: add time-blocked entries and check/uncheck them
- **Track todos**: add, nest (via parent matching), and check/uncheck tasks
- **Fast, local, and private** — your notes stay on your machine
- **Works with your existing Obsidian vault**
- **User-friendly TUI** for smooth, interactive input

## 🏗️ Installation

You can install Onyx in two ways:

### 1. Download a pre-built binary

- Go to the [GitHub Releases page](https://github.com/corani/onyx/releases)
- Download the latest release for your OS (e.g., `onyx-linux-amd64`, `onyx-darwin-amd64`, etc.)
- Make it executable:
   ```bash
   chmod +x ./onyx-<your-platform>
   mv ./onyx-<your-platform> ~/bin/onyx  # or anywhere in your $PATH
   ```

### 2. Install via Go

If you have Go installed (1.21+ recommended):

```bash
go install github.com/corani/onyx/cmd/onyx@latest
```

This will place the `onyx` binary in your `$GOPATH/bin` or `$HOME/go/bin` directory.

## 🚀 Quick Start

1. **Set your Obsidian vault path:**
   
   Create a config file at `$HOME/.config/onyx/config` (or use `$XDG_CONFIG_HOME/onyx/config`) with the following content:

   ```
   ONYX_VAULT=/path/to/your/obsidian/vault
   ```

2. **Add a note for today:**
   
   ```bash
   ./onyx note add "My note for today!"
   ```

3. **List notes for a specific date:**
   
   ```bash
   ./onyx note list -d 2025-10-28
   ```

## 🧩 Prerequisites

Onyx expects your Obsidian vault to have the following structure:

- **Vault location:** Set in your config file as `ONYX_VAULT=/path/to/your/obsidian/vault`.
- **Daily notes location:** Daily notes must be stored at `Journal/Daily/YYYY-MM-DD.md` inside
  your vault.
- **Daily note format:** Each daily note file must contain a section header:

   ```markdown
   ## Notes
   ```

   Onyx will add new notes under this section. If the section is missing, notes cannot be added.

## 📁 How It Works

- Onyx stores your notes as markdown files in your Obsidian vault, under `Journal/Daily/YYYY-MM-DD.md`.
- Each note is added to the `## Notes` section of the daily file.
- You can add notes interactively or directly from the command line.
- Day planner entries live under a `## Day Planner` section in the same daily file.

## 🗓 Day Planner & Todos

Onyx lets you manage scheduled items in a dedicated `## Day Planner` section of your daily note.

### Requirements

Your daily note must already contain:

```markdown
## Day Planner
```

Onyx will not create this section; if it is missing, planner commands return an error.

Planner entries are stored as checkbox list items with a start time and optional end time:

```markdown
- [ ] 09:30 Start work
- [ ] 16:45-17:00 Daily Scrum
```

### Adding Planner Entries

Use a start time (24h `HH:MM`), optional end time, and description. Omitting arguments triggers interactive single-line prompts.

```bash
./onyx plan add 09:30 "Start work"
./onyx plan add 16:45-17:00 "Daily Scrum"
./onyx plan add 09:30          # interactive description prompt
./onyx plan add                # interactive time + description prompts
```

Entries are inserted in chronological order by start time. Overlapping times are allowed; a warning is logged.

### Listing Planner

```bash
./onyx plan list          # today
./onyx plan list -d 2025-11-11
```


### Marking Planner Done / Undone

Provide a case-insensitive substring of the planner entry text (must match exactly one entry):

```bash
./onyx plan check "Start work"
./onyx plan check "Daily Scrum"
./onyx plan uncheck "Start work"
```

### Todo Section

Add a `## Todo` section to your daily note to manage tasks:

```markdown
## Todo
```

Todos are checkbox list items; nesting uses leading tabs:

```markdown
- [ ] Inbox Zero
- [ ] Architecture
   - [ ] Widget generator
- [x] Add 'plan' command to Onyx
```

Add top-level todos:

```bash
./onyx todo add "Inbox Zero"
```

Add a nested todo under a matched parent (case-insensitive substring; must match exactly one):

```bash
./onyx todo add --parent inbox "Archive old mail"
```

Interactive add (omit text) will prompt for the todo text; if `--parent` not supplied it will also
prompt optionally for a parent substring (blank for top-level).

Check / uncheck by substring (must match exactly one todo’s text):

```bash
./onyx todo check "Inbox Zero"
./onyx todo uncheck "Architecture"
```

### Notes

- Time format must be 24h `HH:MM`; end time (if supplied) must be after start.
- Interactive prompts: Enter submits, Esc cancels that field (blank end time allowed).
- Planner section missing → error (no implicit creation).
- Todo section missing → error (no implicit creation).
- A future enhancement may include updating or removing entries.

## 💡 Tips

- Make sure your Obsidian vault already exists before using Onyx.
- Onyx does not create new vaults or daily note files if they are missing.
- All notes are stored locally — no cloud, no accounts, just your notes.

## 🛠️ Need Help?

- If you run into issues, check your `ONYX_VAULT` path and make sure your Obsidian vault is set up.
- For more options, run:
  
  ```bash
  ./onyx --help
  ```

---

Happy journaling! ✍️
