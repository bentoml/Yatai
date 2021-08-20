#!/bin/bash
set -e
set -o pipefail

source ./scripts/ci/ci-helper.sh

echo "Running golangci-lint..."

golangci-lint run

PASS "golangci-lint passed!"
