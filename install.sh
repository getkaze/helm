#!/usr/bin/env bash
# helm installer — https://getkaze.dev/helm/install.sh
# Usage: curl -fsSL https://getkaze.dev/helm/install.sh | bash
set -euo pipefail

BINARY_NAME="helm"
RELEASES_BASE="https://github.com/getkaze/helm/releases"

# Install directory: ~/.local/bin (user-writable).
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"

# ── color helpers ──────────────────────────────────────────────────────────────
bold=$(tput bold 2>/dev/null || true)
reset=$(tput sgr0 2>/dev/null || true)
green=$(tput setaf 2 2>/dev/null || true)
red=$(tput setaf 1 2>/dev/null || true)

info()  { echo "${bold}==>${reset} $*"; }
ok()    { echo "${green}  ✓${reset} $*"; }
fail()  { echo "${red}error:${reset} $*" >&2; exit 1; }

# ── sanity checks ─────────────────────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)  ;;
  darwin) ;;
  *)      fail "unsupported OS: $OS (supported: linux, darwin)" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       fail "unsupported architecture: $ARCH" ;;
esac

# Detect HELM_VERSION (optional — defaults to latest)
VERSION="${HELM_VERSION:-}"
if [ -z "$VERSION" ]; then
  info "fetching latest version"
  VERSION="$(curl -fsSL -o /dev/null -w '%{url_effective}' "${RELEASES_BASE}/latest" 2>/dev/null || true)"
  VERSION="${VERSION##*/}"
  [ -n "$VERSION" ] && [ "$VERSION" != "latest" ] || fail "could not fetch latest version from ${RELEASES_BASE}/latest"
  VERSION="$(echo "$VERSION" | tr -d '[:space:]')"
fi

DOWNLOAD_URL="${RELEASES_BASE}/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"

# ── download binary ────────────────────────────────────────────────────────────
info "installing helm ${VERSION} (${OS}/${ARCH})"

TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

if command -v curl &>/dev/null; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP"
elif command -v wget &>/dev/null; then
  wget -qO "$TMP" "$DOWNLOAD_URL"
else
  fail "curl or wget is required"
fi

chmod +x "$TMP"
mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
ok "binary installed to ${INSTALL_DIR}/${BINARY_NAME}"

# ── done ───────────────────────────────────────────────────────────────────────
echo ""
echo "${bold}helm ${VERSION} installed successfully!${reset}"

# Check if install dir is in PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "${bold}Note:${reset} ${INSTALL_DIR} is not in your PATH."
    echo "  Add it to your shell profile:"
    echo ""
    echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
    ;;
esac

echo ""
echo "Quick start:"
echo "  helm init                # initialize a project"
echo "  helm status              # show pipeline dashboard"
echo "  helm resume              # resume from last session"
echo "  helm save                # checkpoint session state"
echo ""
