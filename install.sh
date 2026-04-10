#!/usr/bin/env sh
set -e

REPO="omriashke/agent-secrets-cli"
BIN_NAME="agent-secrets"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Fetch latest release tag
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine latest release." >&2
  exit 1
fi

FILENAME="${BIN_NAME}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

echo "Installing ${BIN_NAME} ${LATEST} (${OS}/${ARCH})..."
curl -sSL "$URL" -o "/tmp/${BIN_NAME}"
chmod +x "/tmp/${BIN_NAME}"

# Pick install directory: prefer /usr/local/bin if writable, then ~/bin
if [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
echo "${BIN_NAME} installed to ${INSTALL_DIR}/${BIN_NAME}"

# Remind the user to add ~/bin to PATH if needed
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "NOTE: Add ${INSTALL_DIR} to your PATH:"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc  # or ~/.bashrc"
    ;;
esac
