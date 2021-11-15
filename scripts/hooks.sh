#!/bin/sh

GIT_ROOT=$(git rev-parse --show-toplevel)

cd "$GIT_ROOT" || exit

HOOKS_PATH="$GIT_ROOT/.git/hooks"
cd "$HOOKS_PATH" && ln -sf "$GIT_ROOT"/scripts/hooks/* .

