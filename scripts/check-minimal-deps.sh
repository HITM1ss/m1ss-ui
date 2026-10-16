#!/bin/sh
set -eu
. ./build-tags.sh
tags=$(tags_for release)
# Test the executable's graph, not ./..., which also includes retained upstream
# protocol source packages and their tests. A removed import can otherwise
# reintroduce a large engine even when all optional feature tags are absent.
deps=$(go list -deps -tags "$tags" .)
for forbidden in 'github.com/sagernet/cronet-go' 'github.com/sagernet/sing-quic' \
    'github.com/sagernet/quic-go' 'github.com/sagernet/gvisor' \
    'github.com/sagernet/tailscale' 'github.com/sagernet/sing-cloudflared' \
    'github.com/sagernet/sing-openconnect' 'github.com/sagernet/sing-openvpn' \
    'google.golang.org/grpc' 'github.com/caddyserver/certmagic'; do
    if printf '%s\n' "$deps" | grep -F "$forbidden"; then
        echo "Excluded engine is reachable from the shipping executable: $forbidden" >&2
        exit 1
    fi
done
printf '%s\n' "$deps" | grep -q '^github.com/metacubex/utls$'
printf '%s\n' "$deps" | grep -q '^github.com/mattn/go-sqlite3$'
