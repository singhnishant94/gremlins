package arithmetic

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.ArithmeticBase.String(), GetArithmeticBaseMutations)
}

var arithmeticBaseMutations = map[token.Token]token.Token{
	token.ADD: token.SUB,
	token.MUL: token.QUO,
	token.QUO: token.MUL,
	token.REM: token.MUL,
	token.SUB: token.ADD,
}

// MutatorConditionalNegated implements a mutator to improved comparison changes.
func GetArithmeticBaseMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	original := n.Op
	mutated, ok := arithmeticBaseMutations[n.Op]
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
