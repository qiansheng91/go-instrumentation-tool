#!/bin/bash

export TOOL_VERSION=v0.3.0

go install github.com/qiansheng91/go-instrumentation-tool@${TOOL_VERSION}
wget -O /tmp/configuration.yaml "https://github.com/qiansheng91/go-instrumentation-tool/releases/download/${TOOL_VERSION}/configuration.yaml"

echo "back up go mod files"
[[ -f go.mod.bak ]] && cp go.mod go.mod.bak
[[ -f go.sum.bak ]] && cp go.sum go.sum.bak

export INSTRUMENT_CONFIG_FILE=/tmp/configuration.yaml

${GOPATH}/bin/go-instrumentation-tool rewrite-deps .
go build -a -toolexec "${GOPATH}/bin/go-instrumentation-tool wrap" .

echo "restore go mod files"
[[ -f go.mod.bak ]] && cp go.mod.bak go.mod && rm go.mod.bak
[[ -f go.sum.bak ]] && cp go.sum.bak go.sum && rm go.sum.bak
