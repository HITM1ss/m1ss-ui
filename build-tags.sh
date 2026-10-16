#!/bin/sh
# Build tags and the linker flags that go with them, for every pipeline.
#
#   . ./build-tags.sh
#   TAGS=$(tags_for docker)   # assign, so a bad profile aborts under set -e
#   go build -tags "$TAGS" -ldflags "$(ldflags_for docker)" ...
#
# Every shipping profile uses the same VLESS-only registry. with_utls also
# enables sing-box's Reality implementation. SQLite still requires CGO; musl
# is supplied by the Linux C toolchain, not by the unrelated Cronet build tag.
BASE_TAGS="with_utls,m1ss_minimal"

tags_for() {
    case "$1" in
        test|dev|release|windows|docker) echo "$BASE_TAGS" ;;
        *)
            echo "build-tags.sh: unknown build profile '$1'" >&2
            return 1
            ;;
    esac
}

# Only the flags tied to the tags; callers add their own platform linking flags.
ldflags_for() {
    case "$1" in
        test|dev|release|windows|docker) echo "-w -s" ;;
        *)
            echo "build-tags.sh: unknown build profile '$1'" >&2
            return 1
            ;;
    esac
}
