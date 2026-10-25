#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"
GOPROXY=off GOTOOLCHAIN=local go test -count=1 ./...
printf '%s\n' "relay check ok"
