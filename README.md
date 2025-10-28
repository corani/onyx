# Onyx


Onyx is a simple, friendly command-line tool for managing your daily notes in an
[Obsidian](https://obsidian.md/) vault. It helps you quickly add and review notes for any day,
right from your terminal.

## ✨ Features

- **Add notes** to your daily Obsidian journal with a single command
- **List notes** for any date, beautifully rendered in your terminal
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
