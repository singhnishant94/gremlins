package arithmetic

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.InvertBitwise.String(), GetInvertBitwiseMutations)
}

var bitwiseMutations = map[token.Token]token.Token{
	token.AND:     token.OR,
	token.OR:      token.AND,
	token.XOR:     token.AND,
	token.AND_NOT: token.AND,
	token.SHL:     token.SHR,
	token.SHR:     token.SHL,
}

// MutatorConditionalNegated implements a mutator to improved comparison changes.
func GetInvertBitwiseMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	original := n.Op
	mutated, ok := bitwiseMutations[n.Op]
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
