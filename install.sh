#!/bin/bash

set -e

REPO="yetmike/awsctx"
BINARY="awsctx"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
    linux*)     OS=linux;;
    darwin*)    OS=darwin;;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)    ARCH=amd64;;
    aarch64)   ARCH=arm64;;
    arm64)     ARCH=arm64;;
    *)         echo "Unsupported architecture: ${ARCH}"; exit 1;;
esac

echo "Detected ${OS} ${ARCH}..."

# Get latest release tag
echo "Fetching latest version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${LATEST_TAG}" ]; then
    echo "Error: Could not determine latest version."
    exit 1
fi

# Remove 'v' prefix for the filename (if current)
VERSION=${LATEST_TAG#v}

echo "Latest version: ${LATEST_TAG}"

# Asset name format from goreleaser: awsctx_{version}_{os}_{arch}.tar.gz
ASSET_NAME="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET_NAME}"

# Download and Extract
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${DOWNLOAD_URL}..."
curl -sL -o "${TMP_DIR}/archive.tar.gz" "${DOWNLOAD_URL}"

echo "Extracting..."
tar -xzf "${TMP_DIR}/archive.tar.gz" -C "${TMP_DIR}"

# Install
INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
    echo "Installing to ${INSTALL_DIR} requires sudo permissions."
    if command -v sudo >/dev/null 2>&1; then
        echo "Please enter your password if prompted."
        sudo mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        echo "sudo not found. Installing to ~/.local/bin instead."
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "${INSTALL_DIR}"
        mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
        echo "Make sure ${INSTALL_DIR} is in your PATH."
    fi
else
    mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "Successfully installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
"${INSTALL_DIR}/${BINARY}" --version
