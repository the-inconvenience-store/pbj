#!/bin/sh
set -eu

operation=${1:?expected lint, fmt, fmt-check, test, build, or audit}
case "$operation" in
  lint|fmt|fmt-check|test|build|audit) ;;
  *) printf 'Unknown Go check: %s\n' "$operation" >&2; exit 2 ;;
esac

# An initialized module has no packages until implementation starts.
# go list must succeed: compilation and module errors are not empty projects.
packages=$(go list ./...)
if [ -z "$packages" ]; then
  printf '%s: skipped: no Go packages\n' "$operation"
  exit 0
fi

case "$operation" in
  lint) exec golangci-lint run ./... ;;
  fmt) exec golangci-lint fmt ./... ;;
  fmt-check)
    status=0
    diff=$(golangci-lint fmt --diff ./...) || status=$?
    if [ -n "$diff" ]; then
      printf '%s\n' "$diff"
      printf 'Run mise run fmt to apply formatting.\n' >&2
      exit 1
    fi
    exit "$status"
    ;;
  test) exec go test -race ./... ;;
  build) exec go build ./... ;;
  audit) exec govulncheck ./... ;;
esac
