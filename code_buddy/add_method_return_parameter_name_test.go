package code_buddy

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"testing"
)

const ModifyReturnParameterNameTestCaseCategory = "modify_return_parameter_name"

func testModifyMethodReturnParameters(fileName string, t *testing.T) {
	afterModifyCode, expectedCode := readAndModifyTestCaseSourceCode(ModifyReturnParameterNameTestCaseCategory, fileName, func(codeFile *ast.File) {
		ast.Inspect(codeFile, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "instrumentation_method" && x.Type.Results != nil {
					modifyMethodReturnParameters(x)
					return false
				}
			}
			return true
		})
	})

	assert.Equal(t, expectedCode, afterModifyCode)
}

func TestModifyMethodReturnParameters(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "without return parameters",
			code: "without_return_parameter_name",
		},
		{
			name: "without return parameters and the return type is function type",
			code: "without_return_parameter_name_and_return_function_type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testModifyMethodReturnParameters(test.code, t)
		})
	}
}
