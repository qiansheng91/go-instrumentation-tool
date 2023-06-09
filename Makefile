PROJECT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: build_instrumentation_tools
build_instrumentation_tools:
	cd $(PROJECT_DIR) && go install .


.PHONY: run_hello_world_example
run_hello_world_example: build_instrumentation_tools
	cd $(PROJECT_DIR)/examples/helloworld-example && export INSTRUMENT_CONFIG_FILE="./configuration.yaml" && go build -a -x -toolexec "${GOPATH}/bin/go-instrumentation-tool wrap" .


.PHONY: run_gin_example
run_gin_example: build_instrumentation_tools
	cd $(PROJECT_DIR)/examples/gin-example && export INSTRUMENT_CONFIG_FILE="./configuration.yaml" && go build -a -x -toolexec "${GOPATH}/bin/go-instrumentation-tool wrap" .


.PHONY: rewrite_hello_world_example-deps
rewrite_hello_world_example-deps: build_instrumentation_tools
	cd $(PROJECT_DIR)/examples/helloworld-example && export INSTRUMENT_CONFIG_FILE="./configuration.yaml" && ${GOPATH}/bin/go-instrumentation-tool rewrite-deps .


.PHONY: rewrite_gin_example_deps
rewrite_gin_example_deps: build_instrumentation_tools
	cd $(PROJECT_DIR)/examples/gin-example && export INSTRUMENT_CONFIG_FILE="./configuration.yaml" && ${GOPATH}/bin/go-instrumentation-tool rewrite-deps .

.PHONY: build_gin_example
build_gin_example: rewrite_gin_example_deps run_gin_example


.PHONY: rewrite_tchannel-go_example_deps
rewrite_tchannel-go_example_deps: build_instrumentation_tools
	cd $(PROJECT_DIR)/examples/tchannel-example && export INSTRUMENT_CONFIG_FILE="./configuration.yaml" && ${GOPATH}/bin/go-instrumentation-tool rewrite-deps .

.PHONY: build_tchannel-go_example
build_tchannel-go_example: rewrite_tchannel-go_example_deps
	cd $(PROJECT_DIR)/examples/tchannel-example && export INSTRUMENT_CONFIG_FILE="./configuration.yaml" && go build -a -x -toolexec "${GOPATH}/bin/go-instrumentation-tool wrap" .

