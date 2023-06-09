package code_buddy

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

// modifyMethodReturnParameters modifies the return parameters of the method to be named if they are not already named.
func modifyMethodReturnParameters(method *ast.FuncDecl) bool {
	if method.Type.Results == nil {
		log.Printf("Method %s has no return parameters", method.Name.Name)
		return false
	}

	if method.Type.Results.NumFields() == 0 {
		log.Printf("Method %s has no return parameters", method.Name.Name)
		return false
	}

	if len(method.Type.Results.List[0].Names) > 0 {
		log.Printf("Method %s return parameters are already named", method.Name.Name)
		return false
	}

	for i, field := range method.Type.Results.List {
		log.Printf("Method %s return parameters are not named", method.Name.Name)
		field.Names = []*ast.Ident{ast.NewIdent(fmt.Sprintf("ret%d", i))}
	}
	return true
}

// weaveTrackingCode adds the generated code to the method.
func weaveTrackingCode(method *ast.FuncDecl, codeGenerator []TrackingCodeGenerator) {
	stmts := make([]ast.Stmt, 0, len(method.Body.List)+len(codeGenerator))

	for _, playLoad := range codeGenerator {
		stmts = append(stmts, playLoad.Generate())
	}
	stmts = append(stmts, method.Body.List...)
	method.Body.List = stmts
}

// WeaveImport adds the import to the code file. the import path is the path to the instrumentation package.
func WeaveImport(fset *token.FileSet, codeFile *ast.File, name, path string) bool {
	return astutil.AddNamedImport(fset, codeFile, name, path)
}

// WeaveMethod weaves the method with the instrumentation code.
func WeaveMethod(method *ast.FuncDecl, beforeInvocationMethodName, afterInvocationMethodName string) error {
	log.Printf("Weaving method %s", method.Name.Name)
	modifyMethodReturnParameters(method)
	weaveCode := buildTrackingCodeGenerator(method, beforeInvocationMethodName, afterInvocationMethodName)
	weaveTrackingCode(method, weaveCode)
	return nil
}

// buildTrackingCodeGenerator builds the tracking code generator for the method.
func buildTrackingCodeGenerator(method *ast.FuncDecl, instrumentationBeforeMethodName, instrumentationAfterMethodName string) []TrackingCodeGenerator {
	var parameterList []*ast.Field
	var returnList []*ast.Field

	if method.Type.Params != nil {
		parameterList = method.Type.Params.List
	}

	if method.Type.Results != nil {
		returnList = method.Type.Results.List
	}

	return []TrackingCodeGenerator{
		newBeforeMethodCodeGenerator(instrumentationBeforeMethodName, getFieldNames(parameterList)...),
		newAfterMethodCodeGenerator(instrumentationAfterMethodName, getFieldNames(returnList)...),
	}
}

// getFieldNames returns the names of the fields in the parameter.
func getFieldNames(parameter []*ast.Field) []string {
	if len(parameter) == 0 {
		return []string{}
	}

	names := make([]string, 0, len(parameter))
	for _, field := range parameter {
		log.Printf("Field name %s", getFieldName(field))
		names = append(names, getFieldName(field))
	}
	return names
}

// getFieldName returns the name of the field.
func getFieldName(parameter *ast.Field) string {
	if len(parameter.Names) == 0 {
		return ""
	}
	return parameter.Names[0].Name
}
