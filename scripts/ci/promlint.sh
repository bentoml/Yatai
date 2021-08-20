#!/bin/bash
set -e
set -o pipefail

source ./scripts/ci/ci-helper.sh

echo "Running promlint..."

yq r scripts/infra/prometheus/chart/cluster-rule.yaml '"recording_rules.yml"' | promtool check rules /dev/stdin

PASS "promlint passed!"
