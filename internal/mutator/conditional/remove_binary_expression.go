package conditional

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.RemoveBinaryExpression.String(), MutatorRemoveTerm)
}

// MutatorRemoveTerm implements a mutator to remove expression terms.
func MutatorRemoveTerm(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}
	if n.Op != token.LAND && n.Op != token.LOR {
		return nil
	}

	var r *ast.Ident

	switch n.Op {
	case token.LAND:
		r = ast.NewIdent("true")
	case token.LOR:
		r = ast.NewIdent("false")
	}

	x := n.X
	y := n.Y

	return []mutator.Mutation{
		{
			Change: func() {
				n.X = r
			},
			Reset: func() {
				n.X = x
			},
			Pos: n.OpPos,
		},
		{
			Change: func() {
				n.Y = r
			},
			Reset: func() {
				n.Y = y
			},
			Pos: n.OpPos,
		},
	}
}
