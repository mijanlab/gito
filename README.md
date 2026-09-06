<div align="center">

# `gt`

### Git without the syntax headache.

An ultra-fast, keyboard-driven terminal interface with automatic AI commit messages, safe branch navigation, and zero learning curve.

```bash
# Install on macOS & Linux
curl -fsSL https://raw.githubusercontent.com/mijanlab/gito/main/install.sh | sh
```

```powershell
# Install on Windows (PowerShell)
irm https://raw.githubusercontent.com/mijanlab/gito/main/install.ps1 | iex
```

<br />

```text
╭──────────────────────────────────────────────────────────────╮
│  ◈ Gito                                             v0.0.5   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  my-project                                                  │
│  └─ main                                      ● Clean        │
│                                                              │
│  What would you like to do?                                  │
│                                                              │
│  ❯  ↑  Push                                                  │
│     ↓  Pull                                                  │
│     ◇  Commit                                                │
│     ⎇  Branch                                               │
│     ◌  Changes                                               │
│     ◷  History                                               │
│     ✦  AI Settings                                           │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  ↑↓ Navigate   Enter Select   Esc Back   q Quit              │
╰──────────────────────────────────────────────────────────────╯
```

</div>

---

## ⚡ 10-Second Quickstart

Just type **`gt`** inside any project:

```bash
gt
```

That's it. No complicated arguments, no syntax to memorize.

---

## 🚀 Fast CLI Shortcuts

Prefer quick terminal commands? `gt` handles the entire Git lifecycle with automatic AI intelligence:

```bash
# 1. Commit all changes with AI and push in one step
gt push

# 2. Pull newest changes safely without merge accidents
gt pull

# 3. View a clean, human-friendly status of your repository
gt status

# 4. Self-update to the latest version anytime
gt update
```

---

## ✨ Why `gt`?

| What you want to do | Traditional Git | With `gt` |
|---|---|---|
| **Push a new branch** | `git push -u origin feature/login` | `gt push` *(automatically sets upstream)* |
| **Commit with AI** | Write manual message or configure complex hooks | `gt` → `Commit` → `✦ AI Generate` |
| **Pull another branch safely** | `git checkout feat && git pull && git checkout main` | `gt` → `Pull` *(guards against accidental merges)* |
| **Undo / Revert a commit** | `git log` → copy SHA → `git revert <SHA>` | `gt` → `History` → select commit → `Revert` |
| **Review changes** | `git diff` / `git status` | `gt` → `Changes` *(scrollable color diff)* |

---

## 📦 Installation Options

<details open>
<summary><b>🍎 macOS</b></summary>

```bash
# One-line installer (Recommended)
curl -fsSL https://raw.githubusercontent.com/mijanlab/gito/main/install.sh | sh

# Or via Go
go install github.com/mijanlab/gito/cmd/gt@latest
```
</details>

<details open>
<summary><b>🐧 Linux</b></summary>

```bash
# One-line installer (Ubuntu, Debian, Fedora, Arch, etc.)
curl -fsSL https://raw.githubusercontent.com/mijanlab/gito/main/install.sh | sh

# Or via Go
go install github.com/mijanlab/gito/cmd/gt@latest
```
</details>

<details open>
<summary><b>🪟 Windows</b></summary>

```powershell
# Open PowerShell and run:
irm https://raw.githubusercontent.com/mijanlab/gito/main/install.ps1 | iex

# Or via Go
go install github.com/mijanlab/gito/cmd/gt@latest
```
</details>

---

## ⌨️ Keyboard Controls

Navigation in `gt` is built entirely around muscle memory:

| Keys | Description |
|:---:|---|
| **`↑` / `k`** | Move cursor up |
| **`↓` / `j`** | Move cursor down |
| **`Enter`** | Select / Confirm action |
| **`Esc`** | Go back to previous screen / Cancel |
| **`PgUp` / `PgDn`** | Scroll through file diffs |
| **`r`** | Refresh repository state |
| **`q`** | Quit |

---

## ✦ Optional AI Commit Generation

`gt` works **100% offline** with zero configuration required.

If you want automatic AI commit messages, connect it to your local **Ollama** model or any OpenAI-compatible API:

1. Press `✦ AI Settings` in the main menu (or run `gt`).
2. Point it to your preferred model:
   - **Local Ollama** *(Zero-cost, offline)*: `http://localhost:11434/v1` with model `qwen2.5-coder:7b`
   - **LM Studio / vLLM / LocalAI**
   - **OpenAI / OpenRouter / Groq**

---

## 🛠️ Commands Reference

```text
Usage:
  gt [command] [flags]

Commands:
  gt              Open the interactive visual interface
  gt push [-y]    Analyze changes with AI, commit, and push to remote
  gt pull         Pull latest changes safely from upstream
  gt status       Show concise repository state
  gt version      Check current version and available updates
  gt update       Auto-upgrade to the latest release
  gt uninstall    Remove gt from your machine
```

---

## 📄 License

MIT License © 2026 Gito Contributors.
