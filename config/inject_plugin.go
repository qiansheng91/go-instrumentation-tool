package config

import (
	"errors"
	"fmt"
	"github.com/qiansheng91/go-instrumentation-tool/pkg_helper"
	"go/build"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type InjectPlugin interface {
	PluginName() string

	CheckIfPluginIsEnabled() bool

	CheckIfMatchInjectPlugin(pkgName string) TargetPackage

	BuildPluginPackage(string) (CompileInfo, error)

	GetPluginPackageInfo(string) (build.Package, error)

	FetchPluginIfNotExist(string) error

	Imports(string) []string
}

type TargetPackage interface {
	PackagePath() string
	Name() string
	PointCuts() map[string]PointCut
}

type targetPackageImpl struct {
	targetPackage targetPackage
	pointCuts     map[string]PointCut
}

func (t targetPackageImpl) PackagePath() string {
	return t.targetPackage.PackagePath
}

func (t targetPackageImpl) Name() string {
	return t.targetPackage.Name
}

func (t targetPackageImpl) PointCuts() map[string]PointCut {
	return t.pointCuts
}

type CompileInfo interface {
	CompileDir() string
}

type PointCut interface {
	TargetSignature() string
	BeforeAdvice() string
	AfterAdvice() string
}

type injectPluginImpl struct {
	plugin           plugin
	pointCutMappings map[string]targetPackage
}

func (i injectPluginImpl) Imports(buildDir string) []string {
	if err := i.FetchPluginIfNotExist(buildDir); err != nil {
		return nil
	}

	if buildPackage, err := pkg_helper.FetchGoPackageInfo(i.plugin.Plugin_package, buildDir); err != nil {
		return nil
	} else {
		return buildPackage.Imports
	}
}

type compileInfoImpl struct {
	buildFile    string
	importCfgDir string
}

func (c compileInfoImpl) CompileDir() string {
	return c.importCfgDir
}

func (i injectPluginImpl) GetPluginPackageInfo(baseDir string) (build.Package, error) {
	buildDir := path.Join(baseDir, i.PluginName())

	return pkg_helper.FetchGoPackageInfo(i.plugin.Plugin_package, buildDir)
}

func (i injectPluginImpl) FetchPluginIfNotExist(buildDir string) error {
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		err = os.MkdirAll(buildDir, 0777)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(path.Join(buildDir, "go.mod")); os.IsNotExist(err) {
		if _, e := pkg_helper.NewPackageModule("github/qiansheng91/go-auto-instrumentation/plugins", buildDir); e != nil {
			log.Printf("failed to init go module: %s", e)
			return e
		}
	}

	if i.plugin.Path != "" {
		if _, err := os.Stat(i.plugin.Path); os.IsNotExist(err) {
			log.Fatalf("inject package %s does not exist", i.plugin.Path)
			return err
		}

		if _, err := os.Stat(path.Join(i.plugin.Path, "go.mod")); os.IsNotExist(err) {
			log.Fatalf("inject package %s does not have go.mod", i.plugin.Path)
			return err
		}

		absPath, _ := filepath.Abs(i.plugin.Path)
		if _, e := pkg_helper.AddLocalPackage(i.plugin.Plugin_package, absPath, buildDir); e != nil {
			log.Printf("failed to add local package: %s", e)
			return e
		}
	}

	if _, err := pkg_helper.InstallPackage(i.plugin.Plugin_package, buildDir); err != nil {
		log.Printf("failed to fetch package info: %s", err)
		return err
	} else {
		return nil
	}
}

func (i injectPluginImpl) BuildPluginPackage(baseDir string) (CompileInfo, error) {
	buildDir := path.Join(baseDir, i.PluginName())

	if err := i.FetchPluginIfNotExist(buildDir); err != nil {
		return nil, err
	}

	// build plugin and change package
	if buildFile, importCfgDir, err := compilePluginInjectPackage(i.plugin, buildDir); err != nil {
		log.Fatalf("failed to build plugin, err: %v", err)
		return nil, err
	} else {
		return compileInfoImpl{buildFile: buildFile, importCfgDir: importCfgDir}, nil
	}
}

func compilePluginInjectPackage(p plugin, pluginDir string) (buildFile, importCfgDirectory string, err error) {
	var pack build.Package
	if pack, err = pkg_helper.FetchGoPackageInfo(p.Plugin_package, pluginDir); err != nil {
		log.Fatalf("failed to fetch package info, err: %v", err)
		return buildFile, importCfgDirectory, err
	}

	finalPath, _ := os.MkdirTemp("", "")
	buildFile = path.Join(finalPath, fmt.Sprintf("%x", rand.Int())+".a")

	goFiles := make([]string, 0)
	for _, file := range pack.GoFiles {
		goFiles = append(goFiles, path.Join(pack.Dir, file))
	}

	if o, e := pkg_helper.BuildPkg(buildFile, goFiles, pluginDir); e != nil {
		log.Printf("failed to build plugin, err: %v, output: \n%s", e, string(o))
		log.Fatalf("failed to build plugin, err: %v", e)
		return buildFile, importCfgDirectory, e
	} else if !strings.HasPrefix(string(o), "WORK=") {
		return buildFile, importCfgDirectory, errors.New("failed to build plugin, output: \n" + string(o))
	} else {
		importCfgDirectory = strings.TrimPrefix(strings.TrimSpace(string(o)), "WORK=")
		return buildFile, importCfgDirectory, nil
	}
}

func (i injectPluginImpl) CheckIfMatchInjectPlugin(pkgName string) TargetPackage {
	targetP, ok := i.pointCutMappings[pkgName]
	if !ok {
		return nil
	}

	result := make(map[string]PointCut)

	for _, pc := range targetP.PointCuts {
		result[pc.TargetSignature] = newPointCut(pc)
	}

	return &targetPackageImpl{
		targetPackage: targetP,
		pointCuts:     result,
	}
}

func (i injectPluginImpl) GetTargetPackage(name string) TargetPackage {
	targetP := i.pointCutMappings[name]
	result := make(map[string]PointCut)

	for _, pc := range targetP.PointCuts {
		result[pc.TargetSignature] = newPointCut(pc)
	}

	return &targetPackageImpl{
		targetPackage: targetP,
		pointCuts:     result,
	}
}

func (i injectPluginImpl) PluginName() string {
	return i.plugin.Name
}

func (i injectPluginImpl) CheckIfPluginIsEnabled() bool {
	return true
}

func newInjectPluginImpl(p plugin) InjectPlugin {
	pointCutMappings := make(map[string]targetPackage)

	for _, pt := range p.Target_packages {
		pointCutMappings[pt.PackagePath] = pt
	}

	return &injectPluginImpl{
		plugin:           p,
		pointCutMappings: pointCutMappings,
	}
}

type pointCutImpl struct {
	cut pointCut
}

func (p pointCutImpl) BeforeAdvice() string {
	return p.cut.BeforeAdvice
}

func (p pointCutImpl) AfterAdvice() string {
	return p.cut.AfterAdvice
}

func (p pointCutImpl) TargetSignature() string {
	return p.cut.TargetSignature
}

func newPointCut(cut pointCut) PointCut {
	return &pointCutImpl{
		cut: cut,
	}
}
