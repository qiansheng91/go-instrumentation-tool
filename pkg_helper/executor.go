package pkg_helper

import (
	"encoding/json"
	"fmt"
	"go/build"
	"log"
	"os/exec"
	"strings"
)

// execute executes the build command in the given directory.
func execute(buildCmd []string, dir string) ([]byte, error) {
	command := exec.Command("go", buildCmd...)
	command.Dir = dir

	if out, err := command.CombinedOutput(); err != nil {
		log.Printf("failed to execute command: go %s, output: \n%s", buildCmd, string(out))
		return nil, err
	} else {
		return out, err
	}
}

// ModTidy executes `go mod tidy` in the given directory.
func ModTidy(packagePath string) error {
	buildCmd := []string{"mod", "tidy"}
	if _, err := execute(buildCmd, packagePath); err != nil {
		return err
	}
	return nil
}

// FetchGoPackageInfo fetches the package info of the given package name in the given directory.
func FetchGoPackageInfo(packageName, packagePath string) (buildPackage build.Package, err error) {
	buildCmd := []string{"list", "-json", packageName}
	if o, e := execute(buildCmd, packagePath); e != nil {
		return buildPackage, e
	} else {
		if err = json.Unmarshal(o, &buildPackage); err != nil {
			return buildPackage, err
		} else {
			return buildPackage, nil
		}
	}
}

// NewPackageModule initializes a new go module in the given module.
func NewPackageModule(packageName, buildDir string) ([]byte, error) {
	buildCmd := strings.Split(fmt.Sprintf("mod init %s", packageName), " ")
	return execute(buildCmd, buildDir)
}

// AddLocalPackage adds a local package to the given module.
func AddLocalPackage(pkgName, pkgPath, buildDir string) ([]byte, error) {
	buildCmd := strings.Split(fmt.Sprintf("mod edit -replace %s=%s", pkgName, pkgPath), " ")
	return execute(buildCmd, buildDir)
}

// InstallPackage install package dependency
func InstallPackage(pkgName, buildDir string) ([]byte, error) {
	buildCmd := strings.Split(fmt.Sprintf("get -u -t -f %s", pkgName), " ")
	return execute(buildCmd, buildDir)
}

// BuildPkg builds the given go files in the given directory.
func BuildPkg(outputFile string, goFiles []string, buildDir string) ([]byte, error) {
	buildCmd := strings.Split(fmt.Sprintf("build -a -work -o %s %s", outputFile,
		strings.Join(goFiles, " ")), " ")

	return execute(buildCmd, buildDir)
}
