package pkg_helper

type ModuleInfo interface {
	Imports() []string
	Deps() []string
	AddImports(deps []string) error
}

type moduleInfoImpl struct {
	modePath string
	deps     []string
	imports  []string
}

func (m *moduleInfoImpl) Imports() []string {
	if m.imports != nil {
		return m.imports
	}

	if buildPkg, err := FetchGoPackageInfo(".", m.modePath); err != nil {
		return nil
	} else {
		m.imports = buildPkg.Imports
		return buildPkg.Imports
	}
}

func (m *moduleInfoImpl) AddImports(deps []string) error {
	for _, dep := range deps {
		if _, e := InstallPackage(dep, m.modePath); e != nil {
			return nil
		}
	}
	return nil
}

func (m *moduleInfoImpl) Deps() []string {
	if m.deps != nil {
		return m.deps
	}

	if deps, err := FetchGoPackageDeps(".", m.modePath); err != nil {
		return nil
	} else {
		m.deps = deps
		return deps
	}
}

func NewModule(path string) ModuleInfo {
	return &moduleInfoImpl{
		modePath: path,
	}
}
