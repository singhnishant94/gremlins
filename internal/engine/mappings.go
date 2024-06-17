/*
 * Copyright 2022 The Gremlins Authors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package engine

import (
	"go/ast"
	"go/token"

	"github.com/singhnishant94/gremlins/internal/mutator"
)

// tokenMutantType is the mapping from each token.Token and all the
// mutator.Type that can be applied to it.
var tokenMutantType = map[token.Token][]mutator.Type{
	token.ADD:            {mutator.ArithmeticBase},
	token.ADD_ASSIGN:     {mutator.InvertAssignments, mutator.RemoveSelfAssignments},
	token.AND:            {mutator.InvertBitwise},
	token.AND_ASSIGN:     {mutator.RemoveSelfAssignments, mutator.InvertBitwiseAssignments},
	token.AND_NOT:        {mutator.InvertBitwise},
	token.AND_NOT_ASSIGN: {mutator.RemoveSelfAssignments, mutator.InvertBitwiseAssignments},
	token.BREAK:          {mutator.InvertLoopCtrl},
	token.CONTINUE:       {mutator.InvertLoopCtrl},
	token.DEC:            {mutator.IncrementDecrement},
	token.EQL:            {mutator.ConditionalsNegation},
	token.GEQ:            {mutator.ConditionalsBoundary, mutator.ConditionalsNegation},
	token.GTR:            {mutator.ConditionalsBoundary, mutator.ConditionalsNegation},
	token.INC:            {mutator.IncrementDecrement},
	token.LAND:           {mutator.InvertLogical, mutator.RemoveBinaryExpression},
	token.LEQ:            {mutator.ConditionalsBoundary, mutator.ConditionalsNegation},
	token.LOR:            {mutator.InvertLogical, mutator.RemoveBinaryExpression},
	token.LSS:            {mutator.ConditionalsBoundary, mutator.ConditionalsNegation},
	token.MUL:            {mutator.ArithmeticBase},
	token.MUL_ASSIGN:     {mutator.InvertAssignments, mutator.RemoveSelfAssignments},
	token.NEQ:            {mutator.ConditionalsNegation},
	token.OR:             {mutator.InvertBitwise},
	token.OR_ASSIGN:      {mutator.RemoveSelfAssignments, mutator.InvertBitwiseAssignments},
	token.QUO:            {mutator.ArithmeticBase},
	token.QUO_ASSIGN:     {mutator.InvertAssignments, mutator.RemoveSelfAssignments},
	token.REM:            {mutator.ArithmeticBase},
	token.REM_ASSIGN:     {mutator.InvertAssignments, mutator.RemoveSelfAssignments},
	token.SHL:            {mutator.InvertBitwise},
	token.SHL_ASSIGN:     {mutator.RemoveSelfAssignments, mutator.InvertBitwiseAssignments},
	token.SHR:            {mutator.InvertBitwise},
	token.SHR_ASSIGN:     {mutator.RemoveSelfAssignments, mutator.InvertBitwiseAssignments},
	token.SUB:            {mutator.ArithmeticBase},
	token.SUB_ASSIGN:     {mutator.InvertAssignments, mutator.RemoveSelfAssignments},
	token.XOR:            {mutator.InvertBitwise},
	token.XOR_ASSIGN:     {mutator.RemoveSelfAssignments, mutator.InvertBitwiseAssignments},
}

func GetMutantTypes(node ast.Node) []mutator.Type {
	mutatorTypes := []mutator.Type{}
	switch n := node.(type) {
	case *ast.AssignStmt:
		mutatorTypes = append(mutatorTypes, tokenMutantType[n.Tok]...)
	case *ast.BinaryExpr:
		mutatorTypes = append(mutatorTypes, tokenMutantType[n.Op]...)
	case *ast.BranchStmt:
		mutatorTypes = append(mutatorTypes, tokenMutantType[n.Tok]...)
	case *ast.IncDecStmt:
		mutatorTypes = append(mutatorTypes, tokenMutantType[n.Tok]...)
	case *ast.UnaryExpr:
		mutatorTypes = append(mutatorTypes, tokenMutantType[n.Op]...)
	case *ast.BlockStmt:
		mutatorTypes = append(mutatorTypes, mutator.RemoveStatement)
	case *ast.CaseClause:
		mutatorTypes = append(mutatorTypes, mutator.RemoveStatement)
	}
	return mutatorTypes
}
