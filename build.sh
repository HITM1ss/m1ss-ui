#!/bin/sh
set -e

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$ROOT"

if [ ! -f frontend/package.json ]; then
    echo "frontend/package.json is missing. Use a complete checkout of this repository." >&2
    exit 1
fi

(cd frontend && npm ci && npm run build)

echo "Backend"

mkdir -p web/html
rm -fr web/html/*
cp -R frontend/dist/* web/html/

. "$ROOT/build-tags.sh"
TAGS=$(tags_for dev)
LDFLAGS=$(ldflags_for dev)

if [ "$(go env GOOS)" = darwin ]; then
    LDFLAGS="$LDFLAGS -extldflags '-Wl,-no_warn_duplicate_libraries'"
fi
CGO_ENABLED=1 go build -trimpath -ldflags "$LDFLAGS" -tags "$TAGS" -o sui main.go
