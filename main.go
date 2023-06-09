package main

import (
	"github.com/qiansheng91/go-instrumentation-tool/build_wrapper"
	"github.com/qiansheng91/go-instrumentation-tool/deps_helper"
	"log"
	"os"
	"path/filepath"
)

var cfgPath string

func init() {
	var err error
	if cfgPath, err = filepath.Abs(os.Getenv("INSTRUMENT_CONFIG_FILE")); err != nil {
		log.Fatalf("failed to get absolute path of instrumentation config file, err: %v", err)
		panic(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s [wrap|rewrite-deps]", os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {
	case "wrap":
		build_wrapper.Wrap(cfgPath, os.Args[2:])
	case "rewrite-deps":
		deps_helper.RewriteDeps(cfgPath, os.Args[2:])
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
		os.Exit(1)
	}
}
