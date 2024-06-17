package incrementdecrement

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.IncrementDecrement.String(), GetIncDecMutations)
}

var incDecMutations = map[token.Token]token.Token{
	token.DEC: token.INC,
	token.INC: token.DEC,
}

// MutatorConditionalNegated implements a mutator to improved comparison changes.
func GetIncDecMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.IncDecStmt)
	if !ok {
		return nil
	}

	original := n.Tok
	mutated, ok := incDecMutations[n.Tok]
	if !ok {
		return nil
	}

	return []mutator.Mutation{
		{
			Change: func() {
				n.Tok = mutated
			},
			Reset: func() {
				n.Tok = original
			},
			Pos: n.TokPos,
		},
	}
}
