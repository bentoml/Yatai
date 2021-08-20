#!/bin/bash
source ./scripts/ci/ci-helper.sh

echo "Running Go format check..."

set_onfail_callback "FAIL goimports reports errors!"

./scripts/ci/goimports.sh | tee /tmp/fmts

if [ -s /tmp/fmts ]; then
    FAIL "Go format check failed!"
    echo 'Please run `make goimports-fmt` to normalize your code.'
    exit 1
fi

PASS "Go format check passed!"
exit 0
