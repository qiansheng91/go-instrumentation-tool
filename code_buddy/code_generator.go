package code_buddy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"log"
	"strings"
)

const (
	interceptorCodeTemplate = `{{weaveMethod}}([]interface{}{ {{parameters}} })`
)

// TrackingCodeGenerator is an interface that generates code from the expression. The TrackingCodeGenerator generates code that is
// wrapped with statement. The TrackingCodeGenerator is used to generate code that is used to call the method that is used to record the
// method execution.
type TrackingCodeGenerator interface {
	Generate() ast.Stmt
}

// exprCodeGenerator is a TrackingCodeGenerator that generates code from the expression. exprCodeGenerator generates code
// that is wrapped with expression statement. The code is wrapped with expression statement because the code is

type exprCodeGenerator struct {
	code string
}

func (e exprCodeGenerator) ToString() string {
	return e.code
}

func (e exprCodeGenerator) Generate() ast.Stmt {
	payloadExpr, _ := parser.ParseExpr(e.code)
	return &ast.ExprStmt{
		X: payloadExpr,
	}
}

// deferCodeGenerator is a TrackingCodeGenerator that generates code from the expression. deferCodeGenerator generates code
// that is wrapped with defer statement. The code is wrapped with defer statement because the code is executed after

type deferCodeGenerator struct {
	codes []string
}

func (d *deferCodeGenerator) Generate() ast.Stmt {
	smts := make([]ast.Stmt, 0, len(d.codes))
	for _, code := range d.codes {
		p, _ := parser.ParseExpr(code)
		smts = append(smts, &ast.ExprStmt{
			X: p,
		})
	}

	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Type: &ast.FuncType{},
				Body: &ast.BlockStmt{
					List: smts,
				},
			},
		},
	}
}

func (d deferCodeGenerator) ToString() string {
	return "1"
}

func newBeforeMethodCodeGenerator(weaveMethod string, parameterNames ...string) TrackingCodeGenerator {
	return &exprCodeGenerator{
		code: generateInterceptorCodeExpr(weaveMethod, parameterNames),
	}
}

func generateInterceptorCodeExpr(weaveMethod string, parameterNames []string) string {
	var result = interceptorCodeTemplate

	parameters := make(map[string]string)
	parameters["weaveMethod"] = weaveMethod
	parameters["parameters"] = generateArgsStatement(parameterNames)

	for key, value := range parameters {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	log.Printf("generateInterceptorCodeExpr:  %s", result)
	return result
}

func newAfterMethodCodeGenerator(weaveMethod string, returnParameters ...string) TrackingCodeGenerator {
	return &deferCodeGenerator{codes: []string{
		generateInterceptorCodeExpr(weaveMethod, returnParameters),
	}}
}

func newExprCodeGenerator(code string) TrackingCodeGenerator {
	return &exprCodeGenerator{
		code: code,
	}
}

func generateArgsStatement(parameters []string) string {
	if len(parameters) == 0 {
		return ""
	}

	p := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		p = append(p, fmt.Sprintf("&%s", parameter))
	}

	return strings.Join(p, ",")
}
