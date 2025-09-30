package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckerAnalyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "check ",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if name := pass.Pkg.String(); strings.Contains(name, "main") {
		return nil, nil
	}

	containsID := func(exp *ast.CallExpr, name string) bool {
		id, ok := exp.Fun.(*ast.Ident)
		return ok && id.Name == name
	}

	containsFun := func(exp *ast.CallExpr, name string) bool {
		fun, ok := exp.Fun.(*ast.SelectorExpr)
		return ok && fun.Sel.Name == name
	}

	expr := func(x *ast.ExprStmt) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			if containsID(call, "os") && containsFun(call, "Exit") {
				pass.Reportf(x.Pos(), "exit function is not allowed use outside main package")
			}
		}
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ExprStmt:
				expr(x)
			}
			return true
		})
	}

	return nil, nil
}
