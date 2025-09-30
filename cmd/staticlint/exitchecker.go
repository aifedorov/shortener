package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckerAnalyzer is a custom analyzer that prevents the use of os.Exit
// in non-main packages. Direct calls to os.Exit in library code can make
// programs difficult to test and control, as they bypass normal program flow
// and prevent proper cleanup.
//
// The analyzer inspects all function calls and reports an error if it finds
// a direct call to os.Exit() outside of a main package. This helps enforce
// the best practice of using os.Exit only in main functions where it's
// appropriate for final program termination.
var ExitCheckerAnalyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "reports direct calls to os.Exit in non-main packages",
	Run:  run,
}

// run is the main analysis function that examines AST nodes for os.Exit calls.
// It skips main packages and reports any direct calls to os.Exit found in other packages.
//
// The function works by:
//  1. Checking if the current package is a main package - if so, allows os.Exit
//  2. Walking through all AST nodes in the package files
//  3. Looking for expression statements that contain function calls
//  4. Checking if any call matches the pattern os.Exit()
//  5. Reporting violations with their source location
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
