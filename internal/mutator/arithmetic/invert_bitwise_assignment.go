package arithmetic

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.InvertBitwiseAssignments.String(), GetInvertBitwiseAssignmentsMutations)
}

var invertBitwiseAssignmentMutations = map[token.Token]token.Token{
	token.AND_ASSIGN:     token.OR_ASSIGN,
	token.OR_ASSIGN:      token.AND_ASSIGN,
	token.XOR_ASSIGN:     token.AND_ASSIGN,
	token.AND_NOT_ASSIGN: token.AND_ASSIGN,
	token.SHL_ASSIGN:     token.SHR_ASSIGN,
	token.SHR_ASSIGN:     token.SHL_ASSIGN,
}

// MutatorConditionalNegated implements a mutator to improved comparison changes.
func GetInvertBitwiseAssignmentsMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	original := n.Tok
	mutated, ok := invertBitwiseAssignmentMutations[n.Tok]
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
