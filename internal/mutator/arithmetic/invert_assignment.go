package arithmetic

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.InvertAssignments.String(), GetInvertAssignmentsMutations)
}

var invertAssignmentMutations = map[token.Token]token.Token{
	token.ADD_ASSIGN: token.SUB_ASSIGN,
	token.MUL_ASSIGN: token.QUO_ASSIGN,
	token.QUO_ASSIGN: token.MUL_ASSIGN,
	token.REM_ASSIGN: token.REM_ASSIGN,
	token.SUB_ASSIGN: token.ADD_ASSIGN,
}

// MutatorConditionalNegated implements a mutator to improved comparison changes.
func GetInvertAssignmentsMutations(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	original := n.Tok
	mutated, ok := invertAssignmentMutations[n.Tok]
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
