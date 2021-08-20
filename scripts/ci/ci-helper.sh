#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

function PASS() {
    echo -e "$GREEN[PASS]$NC $*"
}

function FAIL() {
    echo -e "$RED[FAIL]$NC $*"
}

function set_onfail_callback() {
    set -E
    trap "$*" ERR
}

set -e
set -o pipefail
