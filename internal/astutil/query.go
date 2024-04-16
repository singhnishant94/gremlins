package astutil

import (
	"go/ast"
)

// IdentifiersInStatement returns all identifiers with their found in a statement.
func IdentifiersInStatement(stmt ast.Stmt) []ast.Expr {
	w := &identifierWalker{}

	ast.Walk(w, stmt)

	return w.identifiers
}

type identifierWalker struct {
	identifiers []ast.Expr
}

func (w *identifierWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.AssignStmt:
		// Handle assignment statements
		for _, expr := range n.Lhs {
			if ident, ok := expr.(*ast.Ident); ok {
				w.identifiers = append(w.identifiers, &ast.Ident{
					Name: ident.Name,
				})

			}
		}
	case *ast.ExprStmt:
		// Handle expression statements
		if call, ok := n.X.(*ast.CallExpr); ok {
			// Example: Check for Call expressions like func calls
			for _, arg := range call.Args {
				if ident, ok := arg.(*ast.Ident); ok {
					w.identifiers = append(w.identifiers, &ast.Ident{
						Name: ident.Name,
					})
				}
			}
		}
	case *ast.IncDecStmt:
		// Handle increment/decrement statements
		if ident, ok := n.X.(*ast.Ident); ok {
			w.identifiers = append(w.identifiers, &ast.Ident{
				Name: ident.Name,
			})
		}
	}

	return w
}

// Functions returns all found functions.
func Functions(n ast.Node) []*ast.FuncDecl {
	w := &functionWalker{}

	ast.Walk(w, n)

	return w.functions
}

type functionWalker struct {
	functions []*ast.FuncDecl
}

func (w *functionWalker) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		w.functions = append(w.functions, n)

		return nil
	}

	return w
}
