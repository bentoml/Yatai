#!/bin/bash
source ./scripts/ci/ci-helper.sh

echo "Running Go format check..."

set_onfail_callback "FAIL gofmt reports errors!"

./scripts/ci/gofmt.sh | tee /tmp/gofmts

if [ -s /tmp/gofmts ]; then
    FAIL "Go format check failed!"
    # shellcheck disable=SC2016
    echo 'Please run `make gofmt-fmt` to normalize your code.'
    exit 1
fi

PASS "Go format check passed!"
exit 0
