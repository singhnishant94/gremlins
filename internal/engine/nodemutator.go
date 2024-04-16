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
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/singhnishant94/gremlins/internal/astutil"
	"github.com/singhnishant94/gremlins/internal/mutator"
)

// StmtRemover is a mutator.Mutator of a ast.Node.
//
// Since the AST is shared among mutants, it is important to avoid that more
// than one mutation is applied to the same file before writing it. For this
// reason, StmtRemover contains a cache of locks, one for each file.
// Every time a mutation is about to being applied, a lock is acquired for
// the file it is operating on. Once the file is written and the token is
// rolled back, the lock is released.
// Keeping a lock per file instead of a lock per StmtRemover allows to apply
// mutations on different files in parallel.
type StmtRemover struct {
	pkgName     string
	fs          *token.FileSet
	file        *ast.File
	node        *Node
	workDir     string
	origFile    []byte
	mutantType  mutator.Type
	status      mutator.Status
	idx         int
	pos         token.Pos
	diff        string
	testExecErr error
}

// NewTokenMutant initialises a NodeMutator.
func NewStmtRemover(
	pkgName string,
	set *token.FileSet,
	file *ast.File,
	node *Node,
	stmtIdx int,
	pos token.Pos,
) *StmtRemover {
	return &StmtRemover{
		pkgName: pkgName,
		fs:      set,
		file:    file,
		node:    node,
		idx:     stmtIdx,
		pos:     pos,
	}
}

// Type returns the mutator.Type of the mutant.Mutator.
func (m *StmtRemover) Type() mutator.Type {
	return m.mutantType
}

// SetType sets the mutator.Type of the mutant.Mutator.
func (m *StmtRemover) SetType(t mutator.Type) {
	m.mutantType = t
}

// Status returns the mutator.Status of the mutant.Mutator.
func (m *StmtRemover) Status() mutator.Status {
	return m.status
}

// SetStatus sets the mutator.Status of the mutant.Mutator.
func (m *StmtRemover) SetStatus(s mutator.Status) {
	m.status = s
}

// Position returns the token.Position where the NodeMutator resides.
func (m *StmtRemover) Position() token.Position {
	return m.fs.Position(m.Pos())
}

// Pos returns the token.Pos where the NodeMutator resides.
func (m *StmtRemover) Pos() token.Pos {
	return m.pos
}

// Diff returns the diff between the original and the mutation.
func (m *StmtRemover) Diff() string {
	return m.diff
}

// SetDiff sets the diff between the original and the mutation.
func (m *StmtRemover) SetDiff(d string) {
	m.diff = d
}

// Pkg returns the package name to which the mutant belongs.
func (m *StmtRemover) Pkg() string {
	return m.pkgName
}

func (m *StmtRemover) Apply() error {
	fileLock(m.Position().Filename).Lock()
	defer fileLock(m.Position().Filename).Unlock()

	filename := filepath.Join(m.workDir, m.fs.Position((*m.node.node).Pos()).Filename)
	var err error
	m.origFile, err = os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Create a copy of the original file to calculate the diff later.
	copyOrigFileName := filepath.Join(m.workDir, m.Position().Filename+".copy.orig")
	if err = m.writeMutatedFile(copyOrigFileName); err != nil {
		return err
	}

	// Statement block removal
	var l []ast.Stmt

	switch n := (*m.node.node).(type) {
	case *ast.BlockStmt:
		l = n.List
	case *ast.CaseClause:
		l = n.Body
	}

	var oldStmt ast.Stmt

	for i := range l {
		if i == m.idx {
			m.pos = l[i].Pos()
			oldStmt = l[i]
			l[i] = astutil.CreateNoopOfStatement(oldStmt)
			break
		}
	}

	if oldStmt == nil {
		fmt.Println("OldStmt is nil. Returning.")
		return nil
	}

	if err = m.writeMutatedFile(filename); err != nil {
		return err
	}

	// Rollback
	for i := range l {
		if i == m.idx {
			l[i] = oldStmt
			break
		}
	}

	m.SetDiff(m.calcDiff(copyOrigFileName, filename))

	// Remove the copy of the original file.
	os.Remove(copyOrigFileName)

	return nil
}

func (m *StmtRemover) writeMutatedFile(filename string) error {
	w := &bytes.Buffer{}
	err := printer.Fprint(w, m.fs, m.file)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, w.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (m *StmtRemover) calcDiff(origFile, mutationFile string) string {
	diff, err := exec.Command("diff", "--label=Original", "--label=New", "-u", origFile, mutationFile).CombinedOutput()
	var execExitCode int
	if err == nil {
		execExitCode = 0
	} else if e, ok := err.(*exec.ExitError); ok {
		execExitCode = e.Sys().(syscall.WaitStatus).ExitStatus()
	} else {
		panic(err)
	}
	if execExitCode != 0 && execExitCode != 1 {
		fmt.Printf("%s\n", diff)

		panic("Could not execute diff on mutation file")
	}

	return string(diff)
}

// Rollback puts back the original file after the test and cleans up the
// NodeMutator to free memory.
func (m *StmtRemover) Rollback() error {
	defer m.resetOrigFile()
	filename := filepath.Join(m.workDir, m.Position().Filename)

	return os.WriteFile(filename, m.origFile, 0600)
}

func (m *StmtRemover) SetTestExecutionError(err error) {
	m.testExecErr = err
}

func (m *StmtRemover) TestExecutionError() error {
	return m.testExecErr
}

// SetWorkdir sets the base path on which to Apply and Rollback operations.
//
// By default, NodeMutator will operate on the same source on which the analysis
// was performed. Changing the workdir will prevent the modifications of the
// original files.
func (m *StmtRemover) SetWorkdir(path string) {
	m.workDir = path
}

// Workdir returns the current working dir in which the Mutator will apply its mutations.
func (m *StmtRemover) Workdir() string {
	return m.workDir
}

func (m *StmtRemover) resetOrigFile() {
	var zeroByte []byte
	m.origFile = zeroByte
}
