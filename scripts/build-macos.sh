#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="$ROOT/dist/macos"
ARCH="${1:-arm64}"
PDFBOX_DIST="$DIST/tools/pdfbox"
export GOCACHE="$ROOT/.gocache"

mkdir -p "$DIST"
mkdir -p "$PDFBOX_DIST"
mkdir -p "$GOCACHE"

go run ./tools/icon/genicon.go

CGO_ENABLED=1 GOOS=darwin GOARCH="$ARCH" go build \
  -trimpath \
  -ldflags "-s -w" \
  -o "$DIST/pdf-splitter" \
  ./cmd/pdfsplitter

if [[ -f "$ROOT/tools/pdfbox/pdfbox-app-3.0.7.jar" ]]; then
  cp "$ROOT/tools/pdfbox/pdfbox-app-3.0.7.jar" "$PDFBOX_DIST/"
elif [[ -f "$ROOT/tools/pdfbox/pdfbox-app-3.0.3.jar" ]]; then
  cp "$ROOT/tools/pdfbox/pdfbox-app-3.0.3.jar" "$PDFBOX_DIST/"
fi

echo "Built: $DIST/pdf-splitter"
