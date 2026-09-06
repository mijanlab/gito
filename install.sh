#!/usr/bin/env sh
set -e

# Gito / gt Installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/mijanlab/gito/main/install.sh | sh

REPO="mijanlab/gito"
BINARY_NAME="gito"
SHORT_NAME="gt"
INSTALL_DIR="${INSTALL_DIR:-}"

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

printf "${BLUE}${BOLD}◈ Installing gt / Gito — Premium Interactive Git TUI...${NC}\n"

# 1. Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
    darwin) TARGET_OS="darwin" ;;
    linux)  TARGET_OS="linux" ;;
    *)
        printf "${RED}Error: Unsupported operating system '%s'.${NC}\n" "$OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64) TARGET_ARCH="amd64" ;;
    arm64|aarch64) TARGET_ARCH="arm64" ;;
    *)
        printf "${RED}Error: Unsupported architecture '%s'.${NC}\n" "$ARCH"
        exit 1
        ;;
esac

# 2. Determine installation destination
if [ -z "$INSTALL_DIR" ]; then
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi
fi

# 3. Check if local build with Go is possible (e.g. running from source repo)
if [ -f "go.mod" ] && [ -d "cmd/gito" ] && command -v go >/dev/null 2>&1; then
    printf "Building from source using local Go...\n"
    go build -ldflags="-s -w" -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/gito
    go build -ldflags="-s -w" -o "${INSTALL_DIR}/${SHORT_NAME}" ./cmd/gt
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${SHORT_NAME}"
    printf "${GREEN}✓ Gito & gt successfully built and installed to ${INSTALL_DIR}${NC}\n"
else
    # Check if prebuilt binary can be downloaded or if Go is available on host
    if command -v go >/dev/null 2>&1; then
        printf "Installing via 'go install'...\n"
        GOBIN="$INSTALL_DIR" go install github.com/mijanlab/gito/cmd/gito@latest
        GOBIN="$INSTALL_DIR" go install github.com/mijanlab/gito/cmd/gt@latest
        printf "${GREEN}✓ Gito & gt successfully installed to ${INSTALL_DIR}${NC}\n"
    else
        # Fallback to GitHub Releases download
        LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
        if [ -z "$LATEST_TAG" ]; then
            LATEST_TAG="v0.0.5"
        fi

        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/gito-${TARGET_OS}-${TARGET_ARCH}"
        printf "Downloading precompiled binary from %s...\n" "$DOWNLOAD_URL"

        TMP_FILE="$(mktemp)"
        if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" 2>/dev/null; then
            mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
            chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
            ln -sf "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${SHORT_NAME}"
            printf "${GREEN}✓ Successfully downloaded and installed 'gt' & 'gito' to ${INSTALL_DIR}${NC}\n"
        else
            rm -f "$TMP_FILE"
            printf "${RED}Error: Could not download binary release.${NC}\n"
            printf "Please ensure Go is installed, or build from source:\n"
            printf "  git clone https://github.com/%s\n  cd gito && make install\n" "$REPO"
            exit 1
        fi
    fi
fi

# 4. PATH verification warning
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
        printf "\n${BLUE}Note:${NC} '%s' is not in your PATH.\n" "$INSTALL_DIR"
        printf "Add it to your shell configuration (e.g. ~/.zshrc or ~/.bashrc):\n"
        printf "  export PATH=\"\$PATH:%s\"\n\n" "$INSTALL_DIR"
        ;;
esac

printf "${GREEN}${BOLD}Done! Type 'gt' to launch.${NC}\n"
