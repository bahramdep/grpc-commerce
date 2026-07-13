#!/bin/sh
set -eu

required_go="go1.26.5"

if ! command -v go >/dev/null 2>&1; then
  echo "go is not installed; required version: ${required_go}" >&2
  exit 1
fi

actual_go="$(go env GOVERSION)"
if [ "${actual_go}" != "${required_go}" ]; then
  echo "wrong Go version: found ${actual_go}, require ${required_go}" >&2
  exit 1
fi

echo "toolchain OK: ${actual_go}"

