package build_wrapper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/qiansheng91/go-instrumentation-tool/assistant"
	"github.com/qiansheng91/go-instrumentation-tool/code_buddy"
	"github.com/qiansheng91/go-instrumentation-tool/config"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type InjectOutput interface {
	SourceFiles() map[string]string
	PluginFiles() map[string]string

	dependenciesCompileDirs() []string
}

type injectOutputImpl struct {
	sourceFileMappings map[string]string
	pluginOutputs      map[string]*pluginOutputImpl
}

func (i *injectOutputImpl) dependenciesCompileDirs() []string {
	var result []string
	for _, pluginOutput := range i.pluginOutputs {
		result = append(result, pluginOutput.compileDir)
	}
	return result
}

func (i *injectOutputImpl) PluginFiles() map[string]string {
	result := make(map[string]string)
	for _, pluginOutput := range i.pluginOutputs {
		for _, sourceFile := range pluginOutput.pluginSourceFiles {
			result[sourceFile] = sourceFile
		}
	}
	return result
}

func (i *injectOutputImpl) SourceFiles() map[string]string {
	result := make(map[string]string)

	for key, value := range i.sourceFileMappings {
		result[key] = value
	}

	return result
}

func (i *injectOutputImpl) getImportPkgInfo() []string {
	var result []string
	for _, pluginOutput := range i.pluginOutputs {
		result = append(result, pluginOutput.getImportCfgFile())
	}
	return result
}

type pluginOutputImpl struct {
	pluginSourceFiles map[string]string
	compileDir        string
}

func (p *pluginOutputImpl) getCompileDir() string {
	return p.compileDir
}

func (p *pluginOutputImpl) getImportCfgFile() string {
	return path.Join(p.compileDir, "b001", "importcfg")
}

func (i *injectOutputImpl) appendSourceFiles(newSourceFileMappings map[string]string) {
	for key, value := range newSourceFileMappings {
		i.sourceFileMappings[key] = value
	}
}

func (i *pluginOutputImpl) appendSourceFiles(newSourceFileMappings map[string]string) {
	for key, value := range newSourceFileMappings {
		i.pluginSourceFiles[key] = value
	}
}

func (i *pluginOutputImpl) setImportCfgDir(compileDir string) {
	i.compileDir = compileDir
}

func InjectCode(ii config.InstrumentationInfo, ba assistant.BuildArgs) (InjectOutput, error) {
	var injectOutput = &injectOutputImpl{
		sourceFileMappings: make(map[string]string),
		pluginOutputs:      make(map[string]*pluginOutputImpl),
	}

	// TODO: 目前不支持多个插件同时注入一个Package
	for _, p := range ii.Plugins() {
		if !p.CheckIfPluginIsEnabled() {
			continue
		}

		targetPackage := p.CheckIfMatchInjectPlugin(ba.PackageName)

		if targetPackage == nil {
			continue
		}

		log.Printf("injecting code for plugin %s", p.PluginName())
		newSources := code_buddy.AttemptToInjectSourceCodeToMethod(ba.Files, func(methodSignature string) []string {
			if pointCut := targetPackage.PointCuts()[methodSignature]; pointCut == nil {
				return nil
			} else {
				return []string{pointCut.BeforeAdvice(), pointCut.AfterAdvice()}
			}
		})

		if len(newSources) == 0 {
			continue
		}

		a, _ := filepath.Abs(ba.ImportCfg)
		log.Printf("%s => %s", ba.ImportCfg, a)

		// write To File
		if sourceFileMappings, err := writeNewSourcesToFiles(ba.GetWorkspace(), newSources); err != nil {
			log.Printf("failed to write new source file, err: %v", err)
			return nil, err
		} else {
			injectOutput.appendSourceFiles(sourceFileMappings)
		}

		pluginOutput := &pluginOutputImpl{
			pluginSourceFiles: make(map[string]string),
		}

		// build plugin package
		if compileInfo, err := p.BuildPluginPackage(ba.GetWorkspace()); err != nil {
			log.Printf("failed to build plugin package, err: %v", err)
			return nil, err
		} else {
			log.Printf("plugin package build success, import cfg file: %s", compileInfo.CompileDir())
			pluginOutput.setImportCfgDir(compileInfo.CompileDir())
		}

		if pluginPackageInfo, err := p.GetPluginPackageInfo(ba.GetWorkspace()); err != nil {
			return nil, err
		} else {
			files := make([]string, 0)

			for _, source := range pluginPackageInfo.GoFiles {
				files = append(files, path.Join(pluginPackageInfo.Dir, source))
			}

			newPluginInfo := code_buddy.PruningPackageNameAndPackageImport(files, targetPackage.PackagePath(), targetPackage.Name())
			if pluginCfgFiles, e := writeNewSourcesToFiles(ba.GetWorkspace(), newPluginInfo); e != nil {
				log.Printf("failed to write new source file, err: %v", err)
			} else {
				pluginOutput.appendSourceFiles(pluginCfgFiles)
			}
		}

		injectOutput.pluginOutputs[p.PluginName()] = pluginOutput
	}

	if len(injectOutput.pluginOutputs) > 0 {
		if err := mergeCompileCfgFileContent(ba.PackageName, ba.ImportCfg, injectOutput.getImportPkgInfo()); err != nil {
			log.Printf("failed to merge compile cfg file, err: %v", err)
			return nil, err
		}

		d := &data{
			PackageName: ba.PackageName,
			CompileDir:  injectOutput.dependenciesCompileDirs(),
		}

		if da, err := json.Marshal(d); err != nil {
			log.Printf("failed to marshal data, err: %v", err)
			return nil, err
		} else {
			if fileName, err := writeDependenciesDirsToFiles(ba.GetWorkspace(), da); err != nil {
				log.Printf("failed to write dependencies dirs to file, err: %v", err)
				return nil, err
			} else {
				log.Printf("write depencies dirs to file %s => %s", fileName, string(da))
			}
		}
	}

	return injectOutput, nil
}

type data struct {
	PackageName string   `json:"package_name"`
	CompileDir  []string `json:"compile_dir"`
}

func ReadDependenciesPkgInfo(ba assistant.BuildArgs) (map[string]string, error) {
	baseDir := filepath.Dir(filepath.Dir(ba.ImportCfg))
	entities, _ := os.ReadDir(baseDir)

	var dependenciesDirs string
	for _, entity := range entities {
		if !entity.IsDir() {
			continue
		}

		dependenciesDir := path.Join(baseDir, entity.Name(), "dependencies-dirs")
		if _, err := os.Stat(dependenciesDir); os.IsNotExist(err) {
			continue
		} else {
			dependenciesDirs = dependenciesDir
		}
	}

	if dependenciesDirs == "" {
		return nil, nil
	}

	if _, err := os.Stat(dependenciesDirs); os.IsNotExist(err) {
		log.Printf("%s not exist", dependenciesDirs)
		return nil, nil
	}

	d := &data{}
	if da, err := ioutil.ReadFile(dependenciesDirs); err != nil {
		return nil, err
	} else {
		log.Printf("read dependencies dirs from file %s => %s", dependenciesDirs, string(da))
		if e := json.Unmarshal(da, d); e != nil {
			return nil, e
		}
	}

	res := make(map[string]string)
	for _, dir := range d.CompileDir {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Printf("dependency dir %s not exist", dir)
			return nil, err
		}

		if pkgs, e := readAllDependenciesPkgs(dir); e != nil {
			log.Printf("failed to read all dependencies pkgs, err: %v", e)
			return nil, e
		} else {
			for key, value := range pkgs {
				if key != "packagefile "+d.PackageName {
					res[key] = value
				}
			}
		}
	}

	return res, nil
}

func readAllDependenciesPkgs(filePath string) (map[string]string, error) {
	var importCfgFiles = make(map[string]string)

	entities, _ := os.ReadDir(filePath)

	for _, entity := range entities {
		cfgFile := path.Join(filePath, entity.Name(), "importcfg")
		var f *os.File
		var err error

		if f, err = os.OpenFile(cfgFile, os.O_RDWR, 0644); err != nil {
			return nil, errors.New("failed to open import cfg file")
		}

		for key, value := range parseCfgFile(f) {
			importCfgFiles[key] = value
		}
		f.Close()
	}

	return importCfgFiles, nil
}

func closeFile(f *os.File) {
	e := f.Close()
	if e != nil {
		log.Fatalf("failed to close file %s", f.Name())
	}
}

func readDependenciesPkgs(filePath string) (map[string]string, error) {
	var importCfgFiles = make(map[string]string)
	var f *os.File
	var err error
	if f, err = os.OpenFile(filePath, os.O_RDWR, 0644); err != nil {
		return nil, errors.New("failed to open import cfg file. " + filePath)
	}
	defer f.Close()

	for key, value := range parseCfgFile(f) {
		importCfgFiles[key] = value
	}
	return importCfgFiles, nil
}

func parseCfgFile(r io.Reader) map[string]string {
	res := make(map[string]string)
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, "packagefile") {
			continue
		}
		index := strings.Index(line, "=")
		if index < 11 {
			continue
		}
		res[line[:index]] = line[index+1:]
	}
	return res
}

func MergeCfgFileContent(cfg string, dependenciesPkgs map[string]string) (err error) {
	if len(dependenciesPkgs) == 0 {
		return nil
	}

	output, err := mergeCfgPkgs(cfg, dependenciesPkgs)
	if err != nil {
		return err
	}

	return rewriteCfgFileContent(cfg, output)
}

func rewriteCfgFileContent(cfg string, output []string) (err error) {
	var f *os.File
	if f, err = os.OpenFile(cfg, os.O_RDWR, 0644); err != nil {
		return errors.New("failed to open link import cfg file")
	}
	defer closeFile(f)

	w := bufio.NewWriter(f)

	for _, line := range output {
		w.WriteString(line + "\n")
	}
	w.Flush()

	return nil
}

func mergeCfgPkgs(cfg string, dependenciesPkgs map[string]string) ([]string, error) {
	file, err := os.Open(cfg)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	var originalPkgs = make(map[string]string)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "packagefile") {
			index := strings.Index(scanner.Text(), "=")
			if index != -1 {
				originalPkgs[scanner.Text()[:index]] = scanner.Text()[index+1:]
			}
		}
		lines = append(lines, scanner.Text())
	}

	for key, value := range dependenciesPkgs {
		if _, ok := originalPkgs[key]; ok {
			continue
		}
		lines = append(lines, key+"="+value)
	}
	return lines, nil
}

func parseImportCfgFile(file string) map[string]string {
	var data []byte
	var err error
	if data, err = ioutil.ReadFile(file); err != nil {
		log.Printf("failed to read import cfg file %s", file)
		return nil
	}

	res := make(map[string]string)
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, "packagefile") {
			continue
		}
		index := strings.Index(line, "=")
		if index < 11 {
			continue
		}
		res[line[:index]] = line[index+1:]
	}

	return res
}

func mergeCompileCfgFileContent(packageName, cfg string, dependenciesImportFiles []string) (err error) {
	if len(dependenciesImportFiles) == 0 {
		return nil
	}

	dependenciesImports := make(map[string]string)
	for _, importCfgFile := range dependenciesImportFiles {
		var dependenciesPkgs map[string]string
		dependenciesPkgs, err = readDependenciesPkgs(importCfgFile)

		if err != nil {
			return err
		}

		for key, value := range dependenciesPkgs {
			if key == "packagefile "+packageName {
				continue
			}
			dependenciesImports[key] = value
		}
	}

	if err = MergeCfgFileContent(cfg, dependenciesImports); err != nil {
		return err
	}

	return nil
}

func writeNewSourcesToFiles(workspace string, sources map[string][]byte) (map[string]string, error) {
	result := make(map[string]string)

	for key, value := range sources {
		if newSourceFile, err := WriteNewSource(workspace, value); err != nil {
			log.Printf("failed to write new source file, err: %v", err)
			return nil, err
		} else {
			log.Printf("new source file: %s \n%s", newSourceFile, string(value))
			result[key] = newSourceFile
		}
	}

	return result, nil
}

func writeDependenciesDirsToFiles(workspace string, data []byte) (string, error) {
	if f, err := os.Create(path.Join(workspace, "dependencies-dirs")); err != nil {
		return "", err
	} else {
		defer func(f *os.File) {
			e := f.Close()
			if e != nil {
				log.Printf("failed to close file %s", f.Name())
			}
		}(f)

		if _, e := f.Write(data); e != nil {
			return "", e
		}
		return f.Name(), nil
	}
}

func WriteNewSource(workspace string, data []byte) (string, error) {
	if f, err := os.CreateTemp(workspace, "instrumentation_*.go"); err != nil {
		return "", err
	} else {
		defer func(f *os.File) {
			e := f.Close()
			if e != nil {
				log.Printf("failed to close file %s", f.Name())
			}
		}(f)

		if _, e := f.Write(data); e != nil {
			return "", e
		}
		return f.Name(), nil
	}
}
