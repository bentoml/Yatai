#!/usr/bin/env bash

set -e

# check if jq command exists
if ! command -v jq &> /dev/null; then
  arch=$(uname -m)
  # download jq from github by different arch
  if [[ $arch == "x86_64" && $OSTYPE == 'darwin'* ]]; then
    jq_archived_name="gojq_v0.12.9_darwin_amd64"
  elif [[ $arch == "arm64" && $OSTYPE == 'darwin'* ]]; then
    jq_archived_name="gojq_v0.12.9_darwin_arm64"
  elif [[ $arch == "x86_64" && $OSTYPE == 'linux'* ]]; then
    jq_archived_name="gojq_v0.12.9_linux_amd64"
  elif [[ $arch == "aarch64" && $OSTYPE == 'linux'* ]]; then
    jq_archived_name="gojq_v0.12.9_linux_arm64"
  else
    echo "jq command not found, please install it first"
    exit 1
  fi
  echo "📥 downloading jq from github"
  if [[ $OSTYPE == 'darwin'* ]]; then
    curl -sL -o /tmp/yatai-jq.zip "https://github.com/itchyny/gojq/releases/download/v0.12.9/${jq_archived_name}.zip"
    echo "✅ downloaded jq to /tmp/yatai-jq.zip"
    echo "📦 extracting yatai-jq.zip"
    unzip -q /tmp/yatai-jq.zip -d /tmp
  else
    curl -sL -o /tmp/yatai-jq.tar.gz "https://github.com/itchyny/gojq/releases/download/v0.12.9/${jq_archived_name}.tar.gz"
    echo "✅ downloaded jq to /tmp/yatai-jq.tar.gz"
    echo "📦 extracting yatai-jq.tar.gz"
    tar zxf /tmp/yatai-jq.tar.gz -C /tmp
  fi
  echo "✅ extracted jq to /tmp/${jq_archived_name}"
  jq="/tmp/${jq_archived_name}/gojq"
else
  jq=$(which jq)
fi

echo "ℹ️  telepresence version: $(telepresence version)"

echo "⌛ telepresence connecting..."
telepresence connect
echo "✅ telepresence connected"

echo "⌛ building yatai api-server in development mode..."
make build-api-server-dev
echo "✅ built yatai api-server in development mode"

echo "⌛ starting yatai api-server..."
env $(kubectl -n yatai-system get secret yatai-env -o jsonpath='{.data}' | $jq 'to_entries|map("\(.key)=\(.value|@base64d)")|.[]' | xargs) ./bin/api-server serve

# telepresence leave yatai-yatai-system || true
# echo "⌛ telepresence intercepting..."
# telepresence intercept yatai -n yatai-system -p 7777:http
# echo "✅ telepresence intercepted"

# function trap_handler() {
#   echo "🛑 received EXIT, exiting..."
#   echo "⌛ kill yatai api-server..."
#   kill ${api_server_pid}
#   echo "✅ yatai api-server killed"
#   echo "⌛ telepresence leaving..."
#   telepresence leave yatai-yatai-system 2> /dev/null || true
#   echo "✅ telepresence left"
#   exit 0
# }

# trap trap_handler EXIT

# sleep infinity
