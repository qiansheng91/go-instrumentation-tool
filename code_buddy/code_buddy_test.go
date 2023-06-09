package code_buddy

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"testing"
)

func TestAttemptToInjectSourceCodeToMethod_Without_Match_Method(t *testing.T) {
	code := []string{"./test_codes/inject_code/without_match_method.code"}

	result := AttemptToInjectSourceCodeToMethod(code, func(methodSignature string) []string {
		return nil
	})

	assert.Equal(t, 0, len(result))
}

func TestAttemptToInjectSourceCodeToMethod_With_Match_Method(t *testing.T) {
	code := []string{"./test_codes/inject_code/match_method.code"}

	result := AttemptToInjectSourceCodeToMethod(code, func(methodSignature string) []string {
		if methodSignature != "instrumentation_method" {
			return nil
		}
		return []string{"beforeInstrumentationMethod", "afterInstrumentationMethod"}
	})

	assert.Equal(t, 1, len(result))
	assert.NotNil(t, result["./test_codes/inject_code/match_method.code"])

	data, _ := ioutil.ReadFile("./test_codes/inject_code/match_method_expected_code.code")
	assert.Equal(t, string(result["./test_codes/inject_code/match_method.code"]), string(data))
}

func TestPruningPackageNameAndPackageImport(t *testing.T) {

	tests := []struct {
		name        string
		sourceCode  string
		packagePath string
		packageName string
	}{
		{
			name:        "test type cast",
			sourceCode:  "remove_pacakge_import_type_cast",
			packagePath: "github.com/gin-gonic/gin",
			packageName: "gin",
		},
		{
			name:        "test function call",
			sourceCode:  "remove_pacakge_import_function_call",
			packagePath: "fmt",
			packageName: "fmt",
		},
		{
			name:        "change package",
			sourceCode:  "change_package",
			packagePath: "main",
			packageName: "main",
		},
		{
			name:        "test type cast",
			sourceCode:  "remove_pacakge_asia_import",
			packagePath: "github.com/gin-gonic/gin",
			packageName: "gin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCode := path.Join("test_codes", "inject_code", tt.sourceCode+".code")
			actualSourceCode := path.Join("test_codes", "inject_code", tt.sourceCode+"_expected_code.code")

			result := PruningPackageNameAndPackageImport([]string{testCode}, tt.packagePath, tt.packageName)
			assert.Equal(t, 1, len(result))
			assert.NotNil(t, result[testCode])

			data, _ := ioutil.ReadFile(actualSourceCode)
			assert.Equal(t, string(result[testCode]), string(data))
		})
	}
}
