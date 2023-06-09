package code_buddy

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"testing"
)

const (
	generateTrackingCodeFoldName = "generate_tracking_code"
	instrumentationPackageName   = "test"
	beforeMethodName             = "beforeInstrumentationMethod"
	afterMethodName              = "afterInstrumentationMethod"
)

func testGenerateTrackingCodeFromMethod(code string, t *testing.T) {
	afterModifyCode, expectedCode := readAndModifyTestCaseSourceCode(generateTrackingCodeFoldName, code, func(codeFile *ast.File) {
		ast.Inspect(codeFile, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "instrumentation_method" {
					WeaveMethod(x, beforeMethodName, afterMethodName)
					return false
				}
			}
			return true
		})
	})

	assert.Equal(t, expectedCode, afterModifyCode)
}

func TestGenerateTrackingCodeForMethod(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "generate tracking code with result parameter name",
			code: "generate_tracking_code_with_return_parameter_name",
		},
		{
			name: "generate tracking code without result parameter name",
			code: "generate_tracking_code_without_return_parameter_name",
		},
		{
			name: "generate tracking code without return",
			code: "generate_tracking_code_without_return",
		},
		{
			name: "generate tracking code without body",
			code: "generate_tracking_code_without_body",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testGenerateTrackingCodeFromMethod(test.code, t)
		})
	}
}
