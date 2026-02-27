#!/bin/bash
set -e

APP=webplayer
VERSION=1.0.0
OUTPUT=dist

echo "🎬 Building WebPlayer v${VERSION}..."
mkdir -p ${OUTPUT}

# Linux amd64
echo "📦 Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${OUTPUT}/${APP}-linux-amd64 .
echo "   ✅ ${OUTPUT}/${APP}-linux-amd64 ($(du -sh ${OUTPUT}/${APP}-linux-amd64 | cut -f1))"

# Linux arm64
echo "📦 Building linux/arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ${OUTPUT}/${APP}-linux-arm64 .
echo "   ✅ ${OUTPUT}/${APP}-linux-arm64 ($(du -sh ${OUTPUT}/${APP}-linux-arm64 | cut -f1))"

echo ""
echo "🎉 Done! Files in ./${OUTPUT}/"
echo ""
echo "Usage:"
echo "  chmod +x ${OUTPUT}/${APP}-linux-amd64"
echo "  ./${OUTPUT}/${APP}-linux-amd64"
echo "  ./${OUTPUT}/${APP}-linux-amd64 -port 9090"
echo "  ./${OUTPUT}/${APP}-linux-amd64 -port 8888 -data /path/to/data.json"
