package deps_helper

import (
	"github.com/qiansheng91/go-instrumentation-tool/code_buddy"
	"github.com/qiansheng91/go-instrumentation-tool/config"
	"github.com/qiansheng91/go-instrumentation-tool/pkg_helper"
	"log"
	"os"
	"path/filepath"
)

func RewriteDeps(cfgPath string, args []string) (err error) {
	modPath := args[len(args)-1]
	modPkg := pkg_helper.NewModule(modPath)

	var instrumentationInfo config.InstrumentationInfo
	if instrumentationInfo, err = config.Parser(cfgPath); err != nil {
		log.Printf("failed to parse config file, err: %v", err)
		return err
	}

	var plugins = findMatchedPlugin(instrumentationInfo.WeavePackages(), modPkg.Deps())
	if len(plugins) < 0 {
		log.Printf("no plugins found")
		return nil
	} else {
		var matchedPlugins = make([]string, 0)
		for _, plugin := range plugins {
			matchedPlugins = append(matchedPlugins, plugin.PluginName())
		}
		log.Printf("found %d matched plugins: %v ", len(plugins), matchedPlugins)
	}

	pluginPath, _ := os.MkdirTemp("", "go-instrumentation-tool-plugin-*")
	pluginDeps := make([]string, 0)
	for _, plugin := range plugins {
		pluDeps := plugin.Imports(pluginPath)
		pluginDeps = append(pluginDeps, arrayDifference(pluDeps, modPkg.Imports())...)
	}

	if len(pluginDeps) <= 0 {
		log.Printf("no plugin deps found")
		return nil
	}

	if err = modPkg.AddImports(pluginDeps); err != nil {
		return err
	}
	if newSource, e := code_buddy.AttemptToInjectDepsImport(additionalDepsFile, additionalDepFileTemplate, pluginDeps); e != nil {
		return e
	} else {
		var additionDepsFile *os.File
		if additionDepsFile, err = os.Create(filepath.Join(modPath, "addition_deps.go")); err != nil {
			log.Printf("failed to create addition_deps.go, err: %v", err)
		}

		if _, err = additionDepsFile.Write(newSource); err != nil {
			log.Printf("failed to write new source to addition_deps.go, err: %v", err)
			return err
		}
	}

	return pkg_helper.ModTidy(modPath)
}

const additionalDepsFile = "addition_deps.go"

const additionalDepFileTemplate = `package main`

func findMatchedPlugin(weavePkgMappings map[string]config.InjectPlugin, depPkgs []string) (plugins []config.InjectPlugin) {
	depPkgsMap := convertArrayToMapForSearch(depPkgs)
	plugins = make([]config.InjectPlugin, 0)

	for targetPackagePath, plugin := range weavePkgMappings {
		// TODO: remove duplicate plugin
		if _, ok := depPkgsMap[targetPackagePath]; ok {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

func arrayDifference(a, b []string) []string {
	log.Printf("Diff Imports:\npluginDeps: %v\nprojectDeps: %v ", a, b)
	m := make(map[string]bool)
	for _, item := range b {
		m[item] = true
	}

	var result []string
	for _, item := range a {
		if _, ok := m[item]; !ok {
			result = append(result, item)
		}
	}
	return result
}

func convertArrayToMapForSearch(depPkgs []string) map[string]bool {
	depPkgsMap := make(map[string]bool)
	for _, depPkg := range depPkgs {
		depPkgsMap[depPkg] = true
	}
	return depPkgsMap
}
