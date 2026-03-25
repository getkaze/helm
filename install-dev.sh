#!/usr/bin/env bash
# helm dev installer — installs bin/helm locally for testing
# Usage: bash install-dev.sh
set -euo pipefail

BINARY_NAME="helm"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="${SCRIPT_DIR}/bin/${BINARY_NAME}"

INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"

bold=$(tput bold 2>/dev/null || true)
reset=$(tput sgr0 2>/dev/null || true)
green=$(tput setaf 2 2>/dev/null || true)
red=$(tput setaf 1 2>/dev/null || true)

info() { echo "${bold}==>${reset} $*"; }
ok()   { echo "${green}  ✓${reset} $*"; }
fail() { echo "${red}error:${reset} $*" >&2; exit 1; }

[ -f "$BIN_PATH" ] || fail "bin/helm not found — run 'make build' first"

# ── install binary ─────────────────────────────────────────────────────────────
info "installing ${BIN_PATH} -> ${INSTALL_DIR}/${BINARY_NAME}"
cp "$BIN_PATH" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
ok "binary installed"

# ── done ───────────────────────────────────────────────────────────────────────
VERSION="$("${INSTALL_DIR}/${BINARY_NAME}" version 2>/dev/null || echo "dev")"
echo ""
echo "${bold}helm ${VERSION} installed (dev build)${reset}"
echo ""
echo "  helm init       # initialize a project"
echo "  helm status     # show pipeline dashboard"
echo "  helm update     # check for updates"
echo "  helm --help     # all commands"
echo ""
