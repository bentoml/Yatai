#!/bin/bash
set -e
set -o pipefail

source ./scripts/ci/ci-helper.sh

echo "Running spellcheck..."

# shellcheck disable=SC2046
misspell -error $(find manager/ui/src -name '*.ts*' -not -path '*/node_modules/*') $(find . -name '*.go')

PASS "spellcheck passed!"
