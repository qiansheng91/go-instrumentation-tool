package code_buddy

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"testing"
)

const ADD_GENERATE_CODE_TESTCASE_FOLDER_NAME = "add_generate_code"

func testAddGenerateCode(fileName string, t *testing.T) {
	afterModifyCode, expectedCode := readAndModifyTestCaseSourceCode(ADD_GENERATE_CODE_TESTCASE_FOLDER_NAME, fileName, func(codeFile *ast.File) {
		ast.Inspect(codeFile, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "instrumentation_method" && x.Type.Results != nil {
					playLoad := []TrackingCodeGenerator{newExprCodeGenerator("println(\"Hello \")"), newExprCodeGenerator("println(\"World\")")}
					weaveTrackingCode(x, playLoad)
					return false
				}
			}
			return true
		})
	})
	assert.Equal(t, expectedCode, afterModifyCode)
}

func TestAddGenerateCode(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "add generate code",
			code: "add_generate_code",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testAddGenerateCode(test.code, t)
		})
	}
}
