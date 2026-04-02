#!/bin/bash
set -euo pipefail

VERSION="${1:-1.0.0}"
OUTPUT_DIR="./dist"
MODULE="github.com/coah80/coahgpt"

PURPLE='\033[38;2;203;166;247m'
GREEN='\033[38;2;166;227;161m'
BOLD='\033[1m'
RESET='\033[0m'

echo -e "${PURPLE}${BOLD}Building coahGPT v${VERSION}${RESET}\n"

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

LDFLAGS="-s -w -X main.version=${VERSION}"

TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for target in "${TARGETS[@]}"; do
    os="${target%/*}"
    arch="${target#*/}"
    output="${OUTPUT_DIR}/coahgpt_${os}_${arch}"

    echo -e "  ${PURPLE}building${RESET} ${os}/${arch}..."
    GOOS="${os}" GOARCH="${arch}" go build -ldflags="${LDFLAGS}" -o "${output}" ./cmd/coahgpt/

    echo -e "  ${GREEN}done${RESET} → ${output}"
done

echo ""

echo -e "${PURPLE}${BOLD}Building coahgpt-server${RESET}\n"

for target in "linux/amd64" "linux/arm64"; do
    os="${target%/*}"
    arch="${target#*/}"
    output="${OUTPUT_DIR}/coahgpt-server_${os}_${arch}"

    echo -e "  ${PURPLE}building${RESET} ${os}/${arch}..."
    GOOS="${os}" GOARCH="${arch}" go build -ldflags="${LDFLAGS}" -o "${output}" ./cmd/coahgpt-server/

    echo -e "  ${GREEN}done${RESET} → ${output}"
done

echo ""
echo -e "${GREEN}${BOLD}All builds complete!${RESET}"
ls -lh "${OUTPUT_DIR}"/
