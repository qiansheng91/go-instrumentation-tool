package assistant

import "strings"

func ParseBuildArgs(args []string) BuildArgs {
	var name string
	var files []string
	var importCfgFile string

	for i := 0; i < len(args); i++ {
		if args[i] == "-importcfg" {
			importCfgFile = args[i+1]
			i++
		}

		if args[i] == "-p" {
			name = args[i+1]
			i++
		}

		if strings.HasSuffix(args[i], ".go") {
			files = args[i:]
			break
		}
	}
	return BuildArgs{
		PackageName: name,
		Files:       files,
		ImportCfg:   importCfgFile,
	}
}
