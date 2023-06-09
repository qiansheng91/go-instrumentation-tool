package code_buddy

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"strings"
)

func getDeclIdentify(f *ast.FuncDecl) (injectPointExpr string) {
	if f.Recv != nil {
		if t, ok := f.Recv.List[0].Type.(*ast.Ident); ok {
			injectPointExpr += t.Name + "."
		}
	}
	injectPointExpr += f.Name.Name
	return injectPointExpr
}

type MethodFilter func(methodSignature string) []string

func PruningPackageNameAndPackageImport(sourceCodes []string, packagePath, packageName string) map[string][]byte {
	fset := token.NewFileSet()
	newSourceCodeMappings := make(map[string][]byte, 0)

	for _, file := range sourceCodes {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			log.Printf("failed to parse file %s, err: %v", file, err)
			continue
		}

		f.Name.Name = packageName

		var packageAsia string
		if strings.Contains(packagePath, "/") {
			packageAsia = packagePath[strings.LastIndex(packagePath, "/")+1:]
		} else {
			packageAsia = packagePath
		}

		astutil.Apply(f, func(c *astutil.Cursor) bool {
			switch expr := c.Node().(type) {
			case *ast.ImportSpec:
				if expr.Path.Value == "\""+packagePath+"\"" {
					if expr.Name != nil {
						packageAsia = expr.Name.Name
					}
					c.Delete()
				}
			case *ast.SelectorExpr:
				ident, ok := expr.X.(*ast.Ident)
				if !ok {
					return true
				}

				if packageAsia != "" && ident.Name != packageAsia {
					return true
				}

				newExpr := &ast.Ident{
					Name: expr.Sel.Name,
				}
				c.Replace(newExpr)
				return false
			}
			return true
		}, nil)
		if newSourceCode, e := generateNewSourceCodeBytes(fset, f); e != nil {
			log.Fatalf("failed to generate new source code for file %s, err: %v", file, e)
		} else {
			newSourceCodeMappings[file] = newSourceCode
		}
	}

	return newSourceCodeMappings
}

func AttemptToInjectDepsImport(sourceFile, src string, additionalDeps []string) ([]byte, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, sourceFile, src, 0)
	if err != nil {
		log.Printf("failed to parse file %s, err: %v", sourceFile, err)
		return nil, err
	}

	for _, dep := range additionalDeps {
		if !astutil.AddNamedImport(fset, f, "_", dep) {
			log.Printf("failed to add import")
		}
	}

	return generateNewSourceCodeBytes(fset, f)
}

// AttemptToInjectSourceCodeToMethod attempts to inject source code to the method
func AttemptToInjectSourceCodeToMethod(sourceCodes []string, methodFilter MethodFilter) map[string][]byte {
	fset := token.NewFileSet()
	newSourceCodeMappings := make(map[string][]byte, 0)

	for _, file := range sourceCodes {
		matched := false

		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			log.Printf("failed to parse file %s, err: %v", file, err)
			continue
		}

		for _, decl := range f.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				advisers := methodFilter(getDeclIdentify(funcDecl))
				if advisers == nil || len(advisers) != 2 {
					continue
				}

				matched = true
				if e := WeaveMethod(funcDecl, advisers[0], advisers[1]); e != nil {
					log.Fatalf("failed to inject code for file %s, err: %v", file, e)
				}
			}
		}

		if matched {
			if newSourceCode, e := generateNewSourceCodeBytes(fset, f); e != nil {
				log.Fatalf("failed to generate new source code for file %s, err: %v", file, e)
			} else {
				newSourceCodeMappings[file] = newSourceCode
			}

		}
	}

	return newSourceCodeMappings
}

// generateNewSourceCodeBytes generates new source code bytes for the given file.
func generateNewSourceCodeBytes(fset *token.FileSet, code *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, code); err != nil {
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}
}
