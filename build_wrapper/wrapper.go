package build_wrapper

import (
	"github.com/qiansheng91/go-instrumentation-tool/assistant"
	"github.com/qiansheng91/go-instrumentation-tool/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func Wrap(instrumentFile string, args []string) {
	log.SetPrefix("[go auto instrumentation] ")
	toolAbsPath := args[0]
	buildArg := args[1:]
	_, toolName := filepath.Split(toolAbsPath)

	if runtime.GOOS == "windows" {
		toolName = strings.TrimSuffix(toolName, ".exe")
	}

	var buildArgs assistant.BuildArgs
	var instrumentationInfo config.InstrumentationInfo
	var err error

	if toolName == "compile" || toolName == "link" {
		buildArgs = assistant.ParseBuildArgs(buildArg)
		instrumentationInfo, err = config.Parser(instrumentFile)
	}

	if err == nil && toolName == "compile" {
		var injectResult InjectOutput
		if injectResult, err = InjectCode(instrumentationInfo, buildArgs); err != nil {
			log.Fatalf("failed to inject code, err: %v", err)
		}

		buildArg = rebuildArgs(buildArg, injectResult)
	}

	if err == nil && toolName == "link" {
		var dependenciesPkgInfos map[string]string
		if dependenciesPkgInfos, err = ReadDependenciesPkgInfo(buildArgs); err != nil {
			log.Fatalf("failed to inject linker, err: %v", err)
		}

		if len(dependenciesPkgInfos) > 0 {
			buildArg = rebuildLinkArgs(buildArg, dependenciesPkgInfos)
		}
	}

	cmd := exec.Command(toolAbsPath, buildArg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func rebuildLinkArgs(args []string, infos map[string]string) []string {
	var newArgs = make([]string, len(args))
	for i := 0; i < len(args); i++ {
		newArgs[i] = args[i]
		if args[i] == "-importcfg" {
			if err := MergeCfgFileContent(args[i+1], infos); err != nil {
				log.Fatalf("failed to rewrite import cfg file, err: %v", err)
				return args
			}
		}
	}

	return newArgs
}

func rebuildArgs(args []string, injectOutput InjectOutput) []string {
	newArgs := make([]string, 0)

	sourceFiles := injectOutput.SourceFiles()
	if len(sourceFiles) == 0 {
		return args
	}

	var alreadHandlePluginFiles bool
	for i := 0; i < len(args); i++ {
		if strings.HasSuffix(args[i], ".go") {
			if pkg, ok := sourceFiles[args[i]]; ok {
				log.Printf("replace source file: %s => %s", args[i], pkg)
				newArgs = append(newArgs, pkg)

				if !alreadHandlePluginFiles {
					for _, value := range injectOutput.PluginFiles() {
						log.Printf("add compile file: %s", value)
						newArgs = append(newArgs, value)
					}
					alreadHandlePluginFiles = true
				}
				continue
			}
		}

		newArgs = append(newArgs, args[i])
	}

	log.Printf("args: %s", strings.Join(args, " "))
	log.Printf("new args: %s", strings.Join(newArgs, " "))
	return newArgs
}
