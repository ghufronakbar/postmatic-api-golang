#!/usr/bin/env sh

set -e

APP_NAME="postmatic-api"
BUILD_DIR="./.build"
CMD_DIR="./cmd/api"

echo "🔧 Building $APP_NAME..."

# Buat folder build jika belum ada
mkdir -p "$BUILD_DIR"

# Build binary
go build -o "$BUILD_DIR/$APP_NAME" "$CMD_DIR"

echo "✅ Build selesai:"
ls -lh "$BUILD_DIR/$APP_NAME"
