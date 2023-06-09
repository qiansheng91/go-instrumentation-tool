package code_buddy

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path/filepath"
)

const (
	TestCodeFolderName = "test_codes"
	CodeFileSuffix     = "code"
)

func readAndModifyTestCaseSourceCode(category, testCaseFileName string, modifyCode func(f *ast.File)) (modifiedCode, expectedCode string) {
	modifiedCode = parseCode(readCodeFromFile(category, testCaseFileName), modifyCode)
	expectedCode = parseCode(readCodeFromFile(category, testCaseFileName+"_expected_code"), nil)
	return modifiedCode, expectedCode
}

func parseCode(code string, modifyCode func(f *ast.File)) string {
	fset := token.NewFileSet()
	var buffer *bytes.Buffer
	codeFile, _ := parser.ParseFile(fset, "", code, 0)

	if modifyCode != nil {
		modifyCode(codeFile)
	}
	var output []byte
	buffer = bytes.NewBuffer(output)
	printer.Fprint(buffer, fset, codeFile)
	return buffer.String()
}

func readCodeFromFile(category, filename string) string {
	if content, err := ioutil.ReadFile(filepath.Join(".", TestCodeFolderName, category, fmt.Sprintf("%s.%s", filename, CodeFileSuffix))); err != nil {
		return ""
	} else {
		return string(content)
	}
}
