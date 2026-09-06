<div align="center">

# gt ◈ Gito

**The premium interactive terminal interface for Git.**

Type `gt` and manage your Git repository with speed, clarity, and safety.

[Installation](#installation) • [Quick Start](#quick-start) • [Workflows](#core-workflows) • [Keybindings](#keybindings) • [AI Integration](#ai-integration)

</div>

---

## Overview

**gt** (**Gito**) is built for developers who want a fast, keyboard-first, and context-aware interface for Git without remembering obscure CLI flags.

Type `gt` in any terminal:

```bash
gt
```

---

## Core Workflows

- **Push** — Automatic remote tracking detection, one-click `Push & set upstream`, and protected `--force-with-lease` confirmations.
- **Pull** — Clean pulls with explicit guardrails against accidental branch cross-merging.
- **Commit** — Interactive staging review, custom message editor, and AI commit message generation.
- **Branch** — Instant creation (`git switch -c`), branch switching, and safe deletion (`-d`) with unmerged commit warnings.
- **Changes** — Scrollable syntax-colored diff viewer with addition/deletion statistics.
- **History** — Visual commit timeline, detailed commit inspector, and one-click commit reverts.

---

## Installation

### 🍎 macOS & 🐧 Linux

#### Option 1: One-Line Installer (Recommended)
```bash
curl -fsSL https://raw.githubusercontent.com/mijanlab/gito/main/install.sh | sh
```
*Installs both `gt` and `gito` to your PATH.*

#### Option 2: Go Install
```bash
go install github.com/mijanlab/gito/cmd/gt@latest
```

#### Option 3: Homebrew
```bash
brew install mijanlab/tap/gito
```

---

### 🪟 Windows

#### Option 1: One-Line PowerShell (Recommended)
```powershell
irm https://raw.githubusercontent.com/mijanlab/gito/main/install.ps1 | iex
```

#### Option 2: Scoop
```powershell
scoop bucket add mijanlab https://github.com/mijanlab/scoop-bucket
scoop install gito
```

#### Option 3: Go Install
```powershell
go install github.com/mijanlab/gito/cmd/gt@latest
```

---

## Quick Start

Open any repository in your terminal and type:

```bash
gt
```

*(You can also use `gito`)*

---

## Commands

| Command | Description |
|---|---|
| `gt` | Launch interactive Git TUI in current directory |
| `gt version` | Display current version and check for available updates |
| `gt update` | Check for updates and self-upgrade to the latest version |
| `gt uninstall` | Uninstall binary from your system (add `--purge` to delete config) |
| `gt --help` | Show CLI usage and available options |

---

## Keybindings

| Key | Action |
|---|---|
| `↑` / `k` | Navigate up |
| `↓` / `j` | Navigate down |
| `Enter` | Select / Confirm |
| `Esc` | Go back / Cancel |
| `PgUp` / `PgDn` | Scroll diff viewer |
| `r` | Refresh repository state |
| `q` / `Ctrl+C` | Quit |
| `?` | Show help |

---

## AI Integration (Optional)

`gt` includes optional AI commit message generation powered by **any OpenAI-compatible API**.

- **Ollama** (`http://localhost:11434/v1` with `qwen2.5-coder:7b`)
- **LM Studio** / **vLLM** / **LocalAI**
- **OpenAI** / **OpenRouter** / **Groq**

Configure interactively inside the TUI under **AI Settings** (`✦`) or in your configuration file:

```yaml
ai:
  enabled: true
  provider: ollama
  base_url: http://localhost:11434/v1
  api_key: ollama
  model: qwen2.5-coder:7b
```

### Config Locations
- **macOS**: `~/Library/Application Support/gito/config.yaml`
- **Linux**: `~/.config/gito/config.yaml`
- **Windows**: `%APPDATA%\gito\config.yaml`

---

## Build from Source

```bash
git clone https://github.com/mijanlab/gito.git
cd gito

# Build both 'gt' and 'gito' binaries
make build

# Install globally to /usr/local/bin
make install

# Run test suite
make test
```

---

## License

MIT License © 2026 Gito Contributors.
