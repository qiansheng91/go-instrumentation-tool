package build_wrapper

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestMergeCfgFileContent(t *testing.T) {
	importDeps := make(map[string]string)
	importDeps["packagefile bytes"] = "github.com/uber/jaeger-client-go@v2.26.0"
	content, _ := mergeCfgPkgs("test_cfg_files/test_importcfg", importDeps)
	assert.Equal(t, 21, len(content))

	file, _ := os.Create("importcfg-*")
	rewriteCfgFileContent(file.Name(), content)

	actualData, _ := ioutil.ReadFile(file.Name())
	expectedData, _ := ioutil.ReadFile("test_cfg_files/test_importcfg_expected")

	assert.Equal(t, string(expectedData), string(actualData))
}
