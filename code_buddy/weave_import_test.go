package code_buddy

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"testing"
)

func testWeaveImport(t *testing.T, fileName string) {
	modifyCode, expectedCode := readAndModifyTestCaseSourceCode("weave_import_code", fileName, func(codeFile *ast.File) {
		ast.Inspect(codeFile, func(n ast.Node) bool {
			return true
		})
		WeaveImport(nil, codeFile, "test", "test")
	})

	assert.Equal(t, expectedCode, modifyCode)
}

func TestWeaveImport(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "weave import with group",
			code: "weave_import_with_group",
		},
		{
			name: "weave import without group",
			code: "weave_import_without_group",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testWeaveImport(t, test.code)
		})
	}

}
