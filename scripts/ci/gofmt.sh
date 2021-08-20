#!/bin/bash
set -e
set -o pipefail

SOURCE_LIST_FILE=/tmp/gofmt-source-files.list
find . -not \( -path './vendor' -prune \) -name '*.go' -not -name '*.gw.go' -not -name '*.pb.go' > $SOURCE_LIST_FILE

# shellcheck disable=SC2046
gofmt $* -d $(cat $SOURCE_LIST_FILE)
