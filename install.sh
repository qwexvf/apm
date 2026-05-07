#!/usr/bin/env sh
set -e

REPO="qwexvf/ccpm"
BIN="ccpm"
INSTALL_DIR="${CCPM_INSTALL_DIR:-}"

# ── resolve install dir ────────────────────────────────────────────────────────
if [ -z "$INSTALL_DIR" ]; then
  if [ -d "$HOME/.local/bin" ] && echo "$PATH" | grep -q "$HOME/.local/bin"; then
    INSTALL_DIR="$HOME/.local/bin"
  elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

# ── detect OS / arch ──────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  GOOS="linux"   ;;
  Darwin) GOOS="darwin"  ;;
  MINGW*|MSYS*|CYGWIN*) GOOS="windows" ;;
  *)
    echo "unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)  GOARCH="amd64" ;;
  arm64|aarch64) GOARCH="arm64" ;;
  *)
    echo "unsupported arch: $ARCH" >&2
    exit 1
    ;;
esac

# ── fetch latest release tag ──────────────────────────────────────────────────
LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"

if command -v curl >/dev/null 2>&1; then
  TAG=$(curl -fsSL "$LATEST_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
elif command -v wget >/dev/null 2>&1; then
  TAG=$(wget -qO- "$LATEST_URL" | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
else
  echo "curl or wget required" >&2
  exit 1
fi

# ── fall back to go install if no releases yet ────────────────────────────────
if [ -z "$TAG" ]; then
  if command -v go >/dev/null 2>&1; then
    echo "no release found — building from source with go install..."
    GOBIN="$INSTALL_DIR" go install "github.com/${REPO}@latest"
    echo "installed ${BIN} → ${INSTALL_DIR}/${BIN}"
    exit 0
  else
    echo "no release found and go not available" >&2
    echo "install Go from https://go.dev/dl/ then retry" >&2
    exit 1
  fi
fi

# ── download binary ───────────────────────────────────────────────────────────
EXT=""
[ "$GOOS" = "windows" ] && EXT=".exe"

FILENAME="${BIN}_${TAG}_${GOOS}_${GOARCH}${EXT}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${FILENAME}"
TMP="$(mktemp)"

echo "downloading ${BIN} ${TAG} (${GOOS}/${GOARCH})..."

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP"
else
  wget -qO "$TMP" "$DOWNLOAD_URL"
fi

chmod +x "$TMP"

# ── install ───────────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/${BIN}${EXT}"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "$DEST"
else
  sudo mv "$TMP" "$DEST"
fi

echo "installed ${BIN} ${TAG} → ${DEST}"

# ── PATH hint ─────────────────────────────────────────────────────────────────
if ! command -v "$BIN" >/dev/null 2>&1; then
  echo ""
  echo "note: add ${INSTALL_DIR} to your PATH:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
