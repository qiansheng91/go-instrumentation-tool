#!/bin/bash

export TOOL_VERSION=v0.2.0

go install github.com/qiansheng91/go-instrumentation-tool
echo "go-instrumentation-tool installed"

echo "Download configuration.yaml"
curl -o /tmp/configuration.yaml https://github.com/qiansheng91/go-instrumentation-tool/releases/download/${TOOL_VERSION}/configuration.yaml

echo "back up go.mod"
cp go.mod go.mod.bak
cp go.sum go.sum.bak

export INSTRUMENT_CONFIG_FILE=/tmp/configuration.yaml
${GOPATH}/bin/go-instrumentation-tool rewrite-deps .

go build -a -x -toolexec "${GOPATH}/bin/go-instrumentation-tool wrap" .
