<div align="center">

# `gt` ◈ Gito

**The premium interactive terminal interface for Git.**

Simple • Fast • Keyboard-First • Safe by Default • Optional AI

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
  │     ⎇  Branch                                                │
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

## Quickstart

Install globally once in your terminal:

### macOS & Linux
```bash
curl -fsSL https://raw.githubusercontent.com/mijanlab/gito/main/install.sh | sh
```
*Or with Go:*
```bash
go install github.com/mijanlab/gito/cmd/gt@latest
```

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/mijanlab/gito/main/install.ps1 | iex
```
*Or with Go:*
```powershell
go install github.com/mijanlab/gito/cmd/gt@latest
```

<br />

Launch the interactive interface from inside any repository:

```bash
gt
```

> [!TIP]
> You can also run `gito` — both `gt` and `gito` are installed and point to the same binary.

---

## Fast CLI Usage

Run everyday Git workflows directly without memorizing obscure flags:

```bash
# Analyze changes with AI, create commit, and push to remote
gt push

# Auto-accept AI commit message and push immediately
gt push -y

# Pull latest changes safely without accidental merge conflicts
gt pull

# Human-friendly repository and tracking status
gt status

# Upgrade gt to the newest release
gt update

# Cleanly uninstall gt from your machine
gt uninstall
```

### CLI Options & Commands

| Command | Action |
| :--- | :--- |
| `gt` | Open the interactive TUI menu |
| `gt push [-y]` | AI-commit uncommitted changes and push to remote (auto-sets upstream) |
| `gt pull` | Pull newest upstream commits safely |
| `gt status` | Clean status overview of branch, tracking, and modified files |
| `gt version` | Display current version and check for available updates |
| `gt update` | Self-upgrade to the latest release |
| `gt uninstall` | Remove `gt` and `gito` binaries (use `--purge` to delete config) |
| `gt --help` | Display CLI usage and flag reference |

---

## ✨ Why `gt`?

| What you want to do | Traditional Git | With `gt` |
| :--- | :--- | :--- |
| **Push a new branch** | `git push -u origin feature/login` | `gt push` *(sets upstream automatically)* |
| **Commit with AI** | Write manual message or setup hooks | `gt` $\rightarrow$ `Commit` $\rightarrow$ `✦ AI Generate` |
| **Pull another branch safely** | `git checkout feat && git pull && git checkout main` | `gt` $\rightarrow$ `Pull` *(prevents cross-branch merge accidents)* |
| **Undo / Revert a commit** | `git log` $\rightarrow$ copy SHA $\rightarrow$ `git revert <SHA>` | `gt` $\rightarrow$ `History` $\rightarrow$ select commit $\rightarrow$ `Revert` |
| **Review changes** | `git diff` / `git status` | `gt` $\rightarrow$ `Changes` *(scrollable colored diff)* |

---

## ⌨️ Keyboard Controls (Interactive Mode)

| Keys | Action |
| :---: | :--- |
| **`↑` / `k`** | Move cursor up |
| **`↓` / `j`** | Move cursor down |
| **`Enter`** | Select / Confirm |
| **`Esc`** | Go back to previous screen / Cancel |
| **`PgUp` / `PgDn`** | Scroll through diffs and commit lists |
| **`r`** | Refresh repository state |
| **`q` / `Ctrl+C`** | Quit |
| **`?`** | Help |

---

## ✦ Optional AI Configuration

`gt` works **100% offline** out of the box with zero external dependencies.

If you want automatic AI commit message generation, connect it to your local **Ollama** model or any OpenAI-compatible API:

1. Launch `gt` and select `✦ AI Settings` (or edit your config file directly).
2. Point to your provider:
   - **Local Ollama** *(Default, free & offline)*: `http://localhost:11434/v1` with model `qwen2.5-coder:7b`
   - **LM Studio / vLLM / LocalAI**
   - **OpenAI / OpenRouter / Groq**

### Config File Locations
- **macOS**: `~/Library/Application Support/gito/config.yaml`
- **Linux**: `~/.config/gito/config.yaml`
- **Windows**: `%APPDATA%\gito\config.yaml`

```yaml
ai:
  enabled: true
  provider: ollama
  base_url: http://localhost:11434/v1
  api_key: ollama
  model: qwen2.5-coder:7b
```

---

## Build from Source

```bash
git clone https://github.com/mijanlab/gito.git
cd gito

# Build both 'gt' and 'gito' binaries
make build

# Install globally to /usr/local/bin
make install

# Cross-compile for Linux, macOS, and Windows
make cross-compile

# Run test suite
make test
```

---

<div align="center">
  <sub>Built with precision by <a href="https://github.com/mijanlab">@mijanlab</a> • Distributed under the MIT License</sub>
</div>
