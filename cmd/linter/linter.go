// Package main is a linter package that checks for panic usage and log.Fatal, os.Exit
// usages outside of main.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the analyzer implementation.
var Analyzer = &analysis.Analyzer{
	Name:     "exitchecker",
	Doc:      "Reports use of panic() and log.Fatal/os.Exit outside main() in package main.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isMainPackage := pass.Pkg.Name() == "main"

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}

	var currentFunc *ast.FuncDecl

	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			if _, ok := n.(*ast.FuncDecl); ok {
				currentFunc = nil
			}
			return true
		}

		switch x := n.(type) {
		case *ast.FuncDecl:
			currentFunc = x
		case *ast.CallExpr:
			// Check for direct calls to panic
			if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(x.Pos(), "forbidden use of panic()")
				return true
			}

			// Check for selector expressions like log.Fatal or os.Exit
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				xIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				switch xIdent.Name {
				case "log":
					switch sel.Sel.Name {
					case "Fatal", "Fatalf", "Fatalln":
						if isMainPackage && currentFunc != nil && currentFunc.Name.Name == "main" && currentFunc.Recv == nil {
							// OK: inside main()
						} else if isMainPackage {
							pass.Reportf(x.Pos(), "forbidden call to log.%s outside main() in package main", sel.Sel.Name)
						} else {
							pass.Reportf(x.Pos(), "forbidden call to log.%s in non-main package", sel.Sel.Name)
						}
					}
				case "os":
					if sel.Sel.Name == "Exit" {
						if isMainPackage && currentFunc != nil && currentFunc.Name.Name == "main" && currentFunc.Recv == nil {
							// OK: inside main()
						} else if isMainPackage {
							pass.Reportf(x.Pos(), "forbidden call to os.Exit outside main() in package main")
						} else {
							pass.Reportf(x.Pos(), "forbidden call to os.Exit in non-main package")
						}
					}
				}
			}
		}
		return true
	})

	return nil, nil
}
