#!/bin/env sh

set -euo pipefail
readonly VERSION=1_2_6

GOOS=linux GOARCH=arm64 go build -gcflags="all=-N -l"
podman build -t "ghcr.io/ewirch/echoctl:arm64-${VERSION}" .
podman push "ghcr.io/ewirch/echoctl:arm64-${VERSION}"
