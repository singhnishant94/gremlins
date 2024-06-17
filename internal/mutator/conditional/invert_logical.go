package conditional

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.InvertLogical.String(), GetInvertLogicalMutations)
}

var invertLogicalMutations = map[token.Token]token.Token{
	token.LAND: token.LOR,
	token.LOR:  token.LAND,
}

func GetInvertLogicalMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	original := n.Op
	mutated, ok := invertLogicalMutations[n.Op]
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
