#!/usr/bin/env bash

read -p "This script will delete your local minikube cluster. Are you sure? [Y/n] " -n 1 -r
echo    # (optional) move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit
fi

echo "🚨 Do not close this window, always pay attention to this window to enter the password prompt!"

set -ex

minikube delete || true
minikube start --memory 8192 --cpus 8

sudo minikube tunnel

