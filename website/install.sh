#!/bin/sh
# Leafpress installer script
# Usage: curl -fsSL https://leafpress.in/install.sh | sh
#
# Supports: macOS (Intel/ARM), Linux (x86_64/ARM64)
# Tested on: macOS, Ubuntu, Fedora, Arch Linux

set -e

REPO="shivamx96/leafpress"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="leafpress"

# Colors (if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

info() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)  echo "darwin" ;;
        Linux*)   echo "linux" ;;
        *)        error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "curl or wget is required"
    fi
}

# Download file
download() {
    url="$1"
    output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "curl or wget is required"
    fi
}

# Check if running with sudo/root when needed
check_permissions() {
    if [ -w "$INSTALL_DIR" ]; then
        return 0
    fi

    if [ "$(id -u)" -eq 0 ]; then
        return 0
    fi

    return 1
}

# Main installation
main() {
    echo ""
    printf "${GREEN}"
    echo "  _            __                        "
    echo " | | ___  __ _/ _|_ __  _ __ ___  ___ ___ "
    echo " | |/ _ \/ _\` | |_| '_ \| '__/ _ \/ __/ __|"
    echo " | |  __/ (_| |  _| |_) | | |  __/\__ \__ \\"
    echo " |_|\___|\__,_|_| | .__/|_|  \___||___/___/"
    echo "                  |_|                     "
    printf "${NC}"
    echo ""

    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected: ${OS}/${ARCH}"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)

    if [ -z "$VERSION" ]; then
        error "Could not determine latest version"
    fi

    info "Latest version: ${VERSION}"

    # Construct download URL
    # Expected format: leafpress-{version}-{os}-{arch}.tar.gz
    TARBALL="leafpress-${VERSION}-${OS}-${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # Download
    info "Downloading ${TARBALL}..."
    download "$DOWNLOAD_URL" "${TMP_DIR}/${TARBALL}"

    # Extract
    info "Extracting..."
    tar -xzf "${TMP_DIR}/${TARBALL}" -C "$TMP_DIR"

    # Find the binary (might be in root or in a subdirectory)
    BINARY_PATH=$(find "$TMP_DIR" -name "$BINARY_NAME" -type f | head -1)

    if [ -z "$BINARY_PATH" ]; then
        error "Binary not found in archive"
    fi

    chmod +x "$BINARY_PATH"

    # Install
    info "Installing to ${INSTALL_DIR}..."

    if check_permissions; then
        mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Permission denied. Trying with sudo..."
        sudo mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Verify installation
    if command -v leafpress >/dev/null 2>&1; then
        INSTALLED_VERSION=$(leafpress version 2>/dev/null || echo "unknown")
        success "leafpress installed successfully!"
        echo ""
        echo "  Version:  ${INSTALLED_VERSION}"
        echo "  Location: ${INSTALL_DIR}/${BINARY_NAME}"
        echo ""
        echo "Get started:"
        echo "  ${BLUE}leafpress init my-garden${NC}"
        echo "  ${BLUE}cd my-garden${NC}"
        echo "  ${BLUE}leafpress serve${NC}"
        echo ""
    else
        warn "leafpress was installed but is not in PATH"
        echo "Add ${INSTALL_DIR} to your PATH, or run:"
        echo "  ${INSTALL_DIR}/${BINARY_NAME} version"
    fi
}

main "$@"
