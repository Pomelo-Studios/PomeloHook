#!/usr/bin/env sh
set -e

REPO="Pomelo-Studios/PomeloHook"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="pomelo-hook"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  OS="linux"  ;;
  darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

ASSET="${BINARY_NAME}-${OS}-${ARCH}"
RELEASE_URL="https://github.com/${REPO}/releases/latest/download"
TMP_FILE="/tmp/${BINARY_NAME}"

echo "Downloading ${ASSET}..."
curl -fsSL "${RELEASE_URL}/${ASSET}" -o "$TMP_FILE"
chmod +x "$TMP_FILE"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
  sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
"${INSTALL_DIR}/${BINARY_NAME}" --version
