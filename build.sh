#!/bin/bash

APP_NAME="sshs"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "darwin/amd64" # macOS Intel
    "darwin/arm64" # macOS Apple Silicon
)

OUTPUT_DIR="build"
mkdir -p "$OUTPUT_DIR"

# 编译每个平台的二进制文件
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME=$APP_NAME-$GOOS-$GOARCH

    echo "Building $OUTPUT_NAME..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT_DIR/$OUTPUT_NAME
done

echo "Build complete."

