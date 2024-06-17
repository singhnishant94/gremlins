package statement

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/astutil"
	"github.com/singhnishant94/gremlins/internal/mutator"
)

func init() {
	mutator.Register(mutator.RemoveStatement.String(), GetRemoveStatementMutations)
}

func checkRemoveStatement(node ast.Stmt) bool {
	if astutil.IsAridNode(node) {
		return false
	}

	switch n := node.(type) {
	case *ast.AssignStmt:
		if n.Tok != token.DEFINE {
			return true
		}
	case *ast.ExprStmt, *ast.IncDecStmt:
		return true
	}

	return false
}

// MutatorRemoveStatement implements a mutator to remove statements.
func GetRemoveStatementMutations(node ast.Node) []mutator.Mutation {
	var l []ast.Stmt

	switch n := node.(type) {
	case *ast.BlockStmt:
		l = n.List
	case *ast.CaseClause:
		l = n.Body
	}

	var mutations []mutator.Mutation

	for i, ni := range l {
		if checkRemoveStatement(ni) {
			li := i
			old := l[li]

			mutations = append(mutations, mutator.Mutation{
				Change: func() {
					l[li] = astutil.CreateNoopOfStatement(old)
				},
				Reset: func() {
					l[li] = old
				},
				Pos: l[li].Pos(),
			})
		}
	}

	return mutations
}
