package conditional

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.ConditionalsBoundary.String(), GetConditionalBoundaryMutations)
}

var boundaryMutations = map[token.Token]token.Token{
	token.GEQ: token.GTR,
	token.GTR: token.GEQ,
	token.LEQ: token.LSS,
	token.LSS: token.LEQ,
}

// MutatorConditionalNegated implements a mutator to improved comparison changes.
func GetConditionalBoundaryMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	original := n.Op
	mutated, ok := boundaryMutations[n.Op]
	if !ok {
		return nil
	}

	return []mutator.Mutation{
		{
			Change: func() {
				n.Op = mutated
			},
			Reset: func() {
				n.Op = original
			},
			Pos: n.OpPos,
		},
	}
}
