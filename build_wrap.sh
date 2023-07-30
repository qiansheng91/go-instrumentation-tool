#!/bin/bash

export TOOL_VERSION=v0.5.0

go install github.com/qiansheng91/go-instrumentation-tool@${TOOL_VERSION}

if [[ ! -f /tmp/configuration.yaml ]]; then
  wget -O /tmp/configuration.yaml "https://github.com/qiansheng91/go-instrumentation-tool/releases/download/${TOOL_VERSION}/configuration.yaml"
fi

echo "back up go mod files"
[[ -f go.mod.bak ]] && cp go.mod go.mod.bak
[[ -f go.sum.bak ]] && cp go.sum go.sum.bak
[[ -f addition_deps.go ]] && rm addition_deps.go

export INSTRUMENT_CONFIG_FILE=/tmp/configuration.yaml

${GOPATH}/bin/go-instrumentation-tool rewrite-deps .
go build -a -toolexec "${GOPATH}/bin/go-instrumentation-tool wrap" .
if [[ $? -ne 0 ]]; then
  echo "failed to build package, Please check the error message above"
  exit -1
else
  echo "build package successfully"
  [[ -f go.mod.bak ]] && cp go.mod.bak go.mod && rm go.mod.bak
  [[ -f go.sum.bak ]] && cp go.sum.bak go.sum && rm go.sum.bak
  [[ -f addition_deps.go ]] && rm addition_deps.go
fi
