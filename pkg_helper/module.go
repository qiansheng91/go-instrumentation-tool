package pkg_helper

type ModuleInfo interface {
	Imports() []string
	AddImports(deps []string) error
}

type moduleInfoImpl struct {
	modePath string
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

func (m moduleInfoImpl) AddImports(deps []string) error {
	for _, dep := range deps {
		if _, e := InstallPackage(dep, m.modePath); e != nil {
			return nil
		}
	}
	return nil
}

func NewModule(path string) ModuleInfo {
	return &moduleInfoImpl{
		modePath: path,
	}
}
