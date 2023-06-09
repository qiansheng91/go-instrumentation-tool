package assistant

import "path"

type BuildArgs struct {
	PackageName string
	Files       []string
	ImportCfg   string
}

func (b *BuildArgs) GetWorkspace() string {
	return path.Dir(b.ImportCfg)
}
