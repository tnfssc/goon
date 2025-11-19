#!/usr/bin/env sh
# Installs the latest goon binary for Linux or macOS.

set -e

REPO="tnfssc/goon"
INSTALL_DIR=${INSTALL_DIR:-"$HOME/.local/bin"}
BINARY_NAME="goon"
DOWNLOAD_BASE="https://github.com/${REPO}/releases/latest/download"

usage() {
  cat <<'EOF'
Usage: install.sh [options]

Options:
  -d, --dir <path>   Installation directory (default: $HOME/.local/bin)
  -h, --help         Show this help message

Environment variables:
  INSTALL_DIR        Same as --dir
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    -d|--dir)
      if [ -z "${2:-}" ]; then
        echo "error: --dir requires a value" >&2
        exit 1
      fi
      INSTALL_DIR=$2
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "error: unknown option $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [ -z "$INSTALL_DIR" ]; then
  echo "error: install directory cannot be empty" >&2
  exit 1
fi

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command '$1' not found" >&2
    exit 1
  fi
}

need_cmd uname
need_cmd mktemp

if command -v curl >/dev/null 2>&1; then
  DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOADER="wget"
else
  echo "error: either curl or wget is required" >&2
  exit 1
fi

case "$(uname -s)" in
  Linux)
    OS="linux"
    ;;
  Darwin)
    OS="darwin"
    ;;
  *)
    echo "error: unsupported operating system $(uname -s)" >&2
    exit 1
    ;;

esac

case "$(uname -m)" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "error: unsupported architecture $(uname -m)" >&2
    exit 1
    ;;

esac

TMP_FILE=$(mktemp)
trap 'rm -f "$TMP_FILE"' EXIT

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
URL="${DOWNLOAD_BASE}/${ASSET_NAME}"

printf 'Downloading %s...\n' "$URL"
if [ "$DOWNLOADER" = "curl" ]; then
  curl -fL "$URL" -o "$TMP_FILE"
else
  wget -q "$URL" -O "$TMP_FILE"
fi

chmod +x "$TMP_FILE"

if [ ! -d "$INSTALL_DIR" ]; then
  if mkdir -p "$INSTALL_DIR" 2>/dev/null; then
    :
  else
    echo "Creating $INSTALL_DIR requires elevated permissions."
    need_cmd sudo
    sudo mkdir -p "$INSTALL_DIR"
  fi
fi

TARGET="$INSTALL_DIR/$BINARY_NAME"
MOVE_OK=0
if mv "$TMP_FILE" "$TARGET" 2>/dev/null; then
  MOVE_OK=1
else
  if command -v sudo >/dev/null 2>&1; then
    echo "Elevated permissions needed to write to $INSTALL_DIR"
    sudo mv "$TMP_FILE" "$TARGET"
    MOVE_OK=1
  fi
fi

if [ "$MOVE_OK" -ne 1 ]; then
  echo "error: could not move binary into $INSTALL_DIR" >&2
  exit 1
fi

printf 'goon installed to %s\n' "$TARGET"

case ":$PATH:" in
  *:"$INSTALL_DIR":*)
    ;;
  *)
    echo "warning: $INSTALL_DIR is not on your PATH. Add the following line to your shell profile:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    ;;

esac

printf "Run 'goon --help' to get started.\n"