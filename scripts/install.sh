#!/usr/bin/env sh
# install.sh — Download and install the latest gitlens release
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/ithaquaKr/gitlens/main/scripts/install.sh | sh
#   curl -fsSL ... | sh -s -- --no-cgo   (install CGo-free build)
#   curl -fsSL ... | VERSION=v0.2.0 sh   (install specific version)

set -e

REPO="ithaquaKr/gitlens"
BINARY="gitlens"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
NO_CGO=0

# Parse flags
for arg in "$@"; do
  case "$arg" in
    --no-cgo) NO_CGO=1 ;;
  esac
done

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Resolve version
if [ -z "$VERSION" ]; then
  echo "Fetching latest release..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version. Set VERSION= explicitly."
  exit 1
fi

echo "Installing gitlens ${VERSION} (${OS}/${ARCH})"

# Build asset name
if [ "$NO_CGO" = "1" ]; then
  ASSET_NAME="${BINARY}_nocgo_${OS}_${ARCH}"
else
  ASSET_NAME="${BINARY}_${OS}_${ARCH}"
fi

EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}.${EXT}"
TMP_DIR=$(mktemp -d)
ARCHIVE="${TMP_DIR}/${ASSET_NAME}.${EXT}"

echo "Downloading ${DOWNLOAD_URL}..."
curl -fsSL -o "$ARCHIVE" "$DOWNLOAD_URL"

# Extract
echo "Extracting..."
if [ "$EXT" = "tar.gz" ]; then
  tar -xzf "$ARCHIVE" -C "$TMP_DIR"
else
  unzip -q "$ARCHIVE" -d "$TMP_DIR"
fi

# Install
EXTRACTED_BINARY="${TMP_DIR}/${BINARY}"
if [ "$OS" = "windows" ]; then
  EXTRACTED_BINARY="${TMP_DIR}/${BINARY}.exe"
fi

chmod +x "$EXTRACTED_BINARY"

if [ -w "$INSTALL_DIR" ]; then
  mv "$EXTRACTED_BINARY" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (may require sudo)..."
  sudo mv "$EXTRACTED_BINARY" "${INSTALL_DIR}/${BINARY}"
fi

rm -rf "$TMP_DIR"

# Verify
if command -v "$BINARY" >/dev/null 2>&1; then
  echo "Installed successfully: $(command -v ${BINARY})"
  echo "Run 'gitlens configure' to get started."
else
  echo "Installed to ${INSTALL_DIR}/${BINARY}"
  echo "Make sure ${INSTALL_DIR} is in your PATH."
fi
