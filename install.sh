#!/bin/sh

# dotenv-tui installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/jellydn/dotenv-tui/main/install.sh | sh

set -e

# Configuration
REPO="jellydn/dotenv-tui"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="dotenv-tui"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)  OS="linux" ;;
        darwin*) OS="darwin" ;;
        msys*|mingw*|cygwin*) OS="windows" ;;
        *)       error "Unsupported operating system: $OS" ;;
    esac
    echo "$OS"
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)             error "Unsupported architecture: $ARCH" ;;
    esac
    echo "$ARCH"
}

# Get latest release version
get_latest_version() {
    # Try multiple methods to get the latest version
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "curl or wget is required to download the installer"
    fi
    
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version from GitHub"
    fi
    
    echo "$VERSION"
}

# Download file
download() {
    URL="$1"
    OUTPUT="$2"
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "$OUTPUT"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$OUTPUT" "$URL"
    else
        error "curl or wget is required to download files"
    fi
}

# Verify checksum
verify_checksum() {
    BINARY_FILE="$1"
    CHECKSUM_FILE="$2"
    
    if ! command -v sha256sum >/dev/null 2>&1; then
        warn "sha256sum not found, skipping checksum verification"
        return 0
    fi
    
    EXPECTED=$(awk '{print $1}' < "$CHECKSUM_FILE")
    ACTUAL=$(sha256sum "$BINARY_FILE" | awk '{print $1}')
    
    if [ "$EXPECTED" != "$ACTUAL" ]; then
        error "Checksum verification failed"
    fi
    
    info "Checksum verified"
}

# Main installation process
main() {
    info "Installing dotenv-tui..."
    
    # Detect system
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected system: $OS-$ARCH"
    
    # Get latest version
    VERSION=$(get_latest_version)
    info "Latest version: $VERSION"
    
    # Construct download URL
    if [ "$OS" = "windows" ]; then
        BINARY_FILENAME="${BINARY_NAME}-${OS}-${ARCH}.exe"
    else
        BINARY_FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    fi
    
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_FILENAME"
    CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_FILENAME.sha256"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT
    
    info "Downloading binary..."
    download "$DOWNLOAD_URL" "$TMP_DIR/$BINARY_FILENAME"
    
    info "Downloading checksum..."
    download "$CHECKSUM_URL" "$TMP_DIR/$BINARY_FILENAME.sha256"
    
    # Verify checksum
    verify_checksum "$TMP_DIR/$BINARY_FILENAME" "$TMP_DIR/$BINARY_FILENAME.sha256"
    
    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Install binary
    info "Installing to $INSTALL_DIR/$BINARY_NAME..."
    mv "$TMP_DIR/$BINARY_FILENAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Check if install directory is in PATH
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *) warn "$INSTALL_DIR is not in your PATH. Add it by running:"
           echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
           ;;
    esac
    
    info "Installation complete! Run 'dotenv-tui' to get started."
}

# Run main function
main
