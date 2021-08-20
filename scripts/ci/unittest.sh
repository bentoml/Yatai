#!/bin/bash
set -e

workdir=.cover
profile="$workdir/cover.out"
mode=set

generate_cover_data() {
    rm -rf "$workdir"
    mkdir "$workdir"
    sudo mkdir -p /var/log/megvii
    sudo chown -R $(id -u):$(id -g) /var/log/megvii

    govendor list --no-status +local | grep -E "common/utils|common/version|common/logfilter|common/security" |xargs -n 1 -I{} -P $(nproc) \
        bash -c "go test -covermode=$mode -coverprofile=$workdir/\"\$(echo {} | tr / -).cover\" {}"

    echo "mode: $mode" >"$profile"
    grep -h -v "^mode:" "$workdir"/*.cover >>"$profile"
}

echo "Generating cover data"
generate_cover_data

statement_coverage=$(go tool cover -func="$profile" 2>&1 | tail -n 1)
echo "Total coverage:" $statement_coverage

go tool cover -html="$profile" -o coverage.html
rm -r $workdir
