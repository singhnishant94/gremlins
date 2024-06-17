package loop

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.InvertLoopCtrl.String(), MutatorLoopBreak)
}

var breakMutations = map[token.Token]token.Token{
	token.CONTINUE: token.BREAK,
	token.BREAK:    token.CONTINUE,
}

// MutatorLoopBreak implements a mutator to change continue to break and break to continue.
func MutatorLoopBreak(node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BranchStmt)
	if !ok {
		return nil
	}

	original := n.Tok
	mutated, ok := breakMutations[n.Tok]
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
