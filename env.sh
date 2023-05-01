#!/bin/bash

arch=$(uname -m)

if [[ "$arch" == "x86_64" ]]; then
    export GOARCH="amd64"
elif [[ "$arch" == "arm64" ]]; then
    export GOARCH="arm64"
else
    echo "Unsupported architecture: $arch"
    exit 1
fi

export CGO_ENABLED="1"
export GO111MODULE="on"

echo "GOARCH set to $GOARCH"