package astutil

import "go/ast"

var loggerIdentifiers = map[string]bool{
	"log":           true,
	"fmt":           true,
	"slogger":       true,
	"logger":        true,
	"serrormonitor": true,
}

func IsAridNode(node ast.Node) bool {
	if node == nil {
		return true
	}

	// Base case
	switch n := node.(type) {
	case *ast.ExprStmt:
		if isLoggerStmt(node) {
			return true
		}

		return IsAridNode(n.X)
	case *ast.BlockStmt:
		allChildrenArid := true
		for _, s := range n.List {
			if !IsAridNode(s) {
				allChildrenArid = false
				break
			}
		}
		return allChildrenArid
	case *ast.IfStmt:
		return IsAridNode(n.Body) && IsAridNode(n.Else)
	case *ast.CallExpr:
		return IsAridNode(n.Fun)
	case *ast.Ident:
		if n.Obj == nil {
			return true
		}
		if funDecl, ok := (n.Obj.Decl).(*ast.FuncDecl); ok {
			return IsAridNode(funDecl)
		}
	case *ast.FuncDecl:
		return IsAridNode(n.Body)
	case *ast.CaseClause:
		allChildrenArid := true
		for _, s := range n.Body {
			if !IsAridNode(s) {
				allChildrenArid = false
				break
			}
		}
		return allChildrenArid
	}

	return false
}

func isLoggerStmt(es ast.Node) bool {
	firstIdent := ""
	ast.Inspect(es, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			// Since it's a depth first traversal, first identifier should be
			// one of loggerIdentifiers for the statement to be a logger stmt.
			if firstIdent != "" {
				// If we've already found our candidate don't recurse further
				return false
			}
			firstIdent = ident.Name
			return false
		}

		if firstIdent != "" {
			return false
		}
		return true
	})

	_, found := loggerIdentifiers[firstIdent]

	return found
}
