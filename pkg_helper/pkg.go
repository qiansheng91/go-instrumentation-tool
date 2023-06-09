package pkg_helper

type GoPkgInfo interface {
	Imports() []string

	Deps() (map[string]bool, error)
}

type goPkgInfoImpl struct {
	path    string
	pkgName string
}

func (g goPkgInfoImpl) Deps() (map[string]bool, error) {
	panic("implement me")
}

func (g goPkgInfoImpl) Imports() []string {
	panic("implement me")
}

func NewPkgInfo(path, pkgName string) GoPkgInfo {
	return &goPkgInfoImpl{
		path:    path,
		pkgName: pkgName,
	}
}
