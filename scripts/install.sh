#!/bin/bash
set -euo pipefail

BOLD='\033[1m'
DIM='\033[2m'
PURPLE='\033[38;2;203;166;247m'
GREEN='\033[38;2;166;227;161m'
RED='\033[38;2;243;139;168m'
RESET='\033[0m'

echo ""
echo -e "${PURPLE}${BOLD}  coah code installer${RESET}"
echo -e "${DIM}  powered by vibes & an RTX 3060${RESET}"
echo ""

INSTALL_DIR="${COAH_INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="coahgpt"
INSTALLED_NAME="coah"
BASE_URL="https://coahgpt.com/releases"

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)  os="linux" ;;
        Darwin*) os="darwin" ;;
        *)
            echo -e "${RED}error: unsupported operating system $(uname -s)${RESET}"
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)
            echo -e "${RED}error: unsupported architecture $(uname -m)${RESET}"
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

PLATFORM=$(detect_platform)
DOWNLOAD_URL="${BASE_URL}/${BINARY_NAME}_${PLATFORM}"

echo -e "  ${DIM}platform:${RESET}  ${PLATFORM}"
echo -e "  ${DIM}install to:${RESET} ${INSTALL_DIR}/${INSTALLED_NAME}"
echo ""

if command -v curl &>/dev/null; then
    DOWNLOADER="curl -fsSL -o"
elif command -v wget &>/dev/null; then
    DOWNLOADER="wget -qO"
else
    echo -e "${RED}error: curl or wget required${RESET}"
    exit 1
fi

TMP_DIR=$(mktemp -d)
TMP_FILE="${TMP_DIR}/${INSTALLED_NAME}"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo -e "${DIM}downloading coah code...${RESET}"
$DOWNLOADER "${TMP_FILE}" "${DOWNLOAD_URL}" || {
    echo -e "${RED}error: download failed${RESET}"
    echo -e "${DIM}url: ${DOWNLOAD_URL}${RESET}"
    exit 1
}

chmod +x "${TMP_FILE}"

if [ -w "${INSTALL_DIR}" ]; then
    mv "${TMP_FILE}" "${INSTALL_DIR}/${INSTALLED_NAME}"
else
    echo -e "${DIM}need sudo to install to ${INSTALL_DIR}${RESET}"
    sudo mv "${TMP_FILE}" "${INSTALL_DIR}/${INSTALLED_NAME}"
fi

echo ""
echo -e "${GREEN}${BOLD}coah code installed!${RESET}"
echo ""
echo -e "  run ${PURPLE}${BOLD}coah${RESET} to start"
echo ""
