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
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/singhnishant94/gremlins/internal/coverage"
	"github.com/singhnishant94/gremlins/internal/diff"
	"github.com/singhnishant94/gremlins/internal/engine/workerpool"
	"github.com/singhnishant94/gremlins/internal/exclusion"
	"github.com/singhnishant94/gremlins/internal/mutator"
	"github.com/singhnishant94/gremlins/internal/report"

	"github.com/singhnishant94/gremlins/internal/configuration"
	"github.com/singhnishant94/gremlins/internal/gomodule"
)

const detectAridNodes = true

var loggerIdentifiers = map[string]bool{
	"log":     true,
	"fmt":     true,
	"slogger": true,
	"logger":  true,
}

// Engine is the "engine" that performs the mutation testing.
//
// It traverses the AST of the project, finds which TokenMutator can be applied and
// performs the actual mutation testing.
type Engine struct {
	fs       fs.FS
	jDealer  ExecutorDealer
	codeData CodeData
	mutants  []mutator.Mutator
	module   gomodule.GoModule
	logger   report.MutantLogger
}

// CodeData is used to check if the mutant should be executed.
type CodeData struct {
	Cov       coverage.Profile
	Diff      diff.Diff
	Exclusion exclusion.Rules
}

type Comment struct {
	Body string `json:"body"`
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"`
}

// Option for the Engine initialization.
type Option func(m Engine) Engine

// New instantiates an Engine.
//
// It gets a fs.FS on which to perform the analysis, a CodeData to
// check if the mutants are executable and a sets of Option.
func New(mod gomodule.GoModule, codeData CodeData, jDealer ExecutorDealer, opts ...Option) Engine {
	dirFS := os.DirFS(filepath.Join(mod.Root, mod.CallingDir))
	mut := Engine{
		module:   mod,
		jDealer:  jDealer,
		codeData: codeData,
		fs:       dirFS,
		logger:   report.NewLogger(),
	}
	for _, opt := range opts {
		mut = opt(mut)
	}

	return mut
}

// WithDirFs overrides the fs.FS of the module (mainly used for testing purposes).
func WithDirFs(dirFS fs.FS) Option {
	return func(m Engine) Engine {
		m.fs = dirFS

		return m
	}
}

// Run executes the mutation testing.
//
// It walks the fs.FS provided and checks every .go file which is not a test.
// For each file it will scan for tokenMutations and gather all the mutants found.
func (mu *Engine) Run(ctx context.Context) report.Results {
	// mu.mutantStream = make(chan mutator.Mutator)
	// go func() {
	// defer close(mu.mutantStream)
	start := time.Now()
	fmt.Printf("Start parsing files\n")
	_ = fs.WalkDir(mu.fs, ".", func(path string, _ fs.DirEntry, _ error) error {
		isGoCode := filepath.Ext(path) == ".go" && !strings.HasSuffix(path, "_test.go")

		if isGoCode && !mu.codeData.Exclusion.IsFileExcluded(path) {
			mu.runOnFile(path)
		}

		return nil
	})
	// }()
	runnable := 0
	for _, m := range mu.mutants {
		if m.Status() == mutator.Runnable {
			runnable++
		}
	}

	fmt.Printf("Found %d mutations in %f seconds out of which %d is runnable\n",
		len(mu.mutants), time.Since(start).Seconds(), runnable)

	start = time.Now()
	res := mu.executeTests(ctx)
	res.Elapsed = time.Since(start)
	res.Module = mu.module.Name

	return res
}

func (mu *Engine) runOnFile(fileName string) {
	src, _ := mu.fs.Open(fileName)
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, fileName, src, parser.ParseComments)
	// file, _, , err := mu.parseAndTypeCheckFile(fileName)
	if err != nil {
		_ = src.Close()
		fmt.Printf("Error parsing file %s\n err: %s", fileName, err)
		return
	}
	_ = src.Close()

	ast.Inspect(file, func(node ast.Node) bool {
		n, ok := NewTokenNode(node)
		if !ok {
			return true
		}
		if detectAridNodes && isAridNode(node) {
			return false
		}
		mu.findTokenMutations(fileName, set, file, n)

		return true
	})

	ast.Inspect(file, func(node ast.Node) bool {
		n, ok := NewNode(node)
		if !ok {
			return true
		}
		if detectAridNodes && isAridNode(node) {
			return false
		}
		mu.findNodeMutations(fileName, set, file, n)

		return true
	})
}

func (mu *Engine) findTokenMutations(fileName string, set *token.FileSet, file *ast.File, node *NodeToken) {
	mutantTypes, ok := TokenMutantType[node.Tok()]
	if !ok {
		return
	}

	pkg := mu.pkgName(fileName, file.Name.Name)
	for _, mt := range mutantTypes {
		if !configuration.Get[bool](configuration.MutantTypeEnabledKey(mt)) {
			continue
		}
		mutantType := mt
		tm := NewTokenMutant(pkg, set, file, node)
		tm.SetType(mutantType)
		tm.SetStatus(mu.mutationStatus(set.Position(node.TokPos)))

		mu.mutants = append(mu.mutants, tm)
		// mu.mutantStream <- tm
	}
}

func (mu *Engine) pkgName(fileName, fPkg string) string {
	var pkg string
	fn := fmt.Sprintf("%s/%s", mu.module.CallingDir, fileName)
	p := filepath.Dir(fn)
	for {
		if strings.HasSuffix(p, fPkg) {
			pkg = fmt.Sprintf("%s/%s", mu.module.Name, p)

			break
		}
		d := filepath.Dir(p)
		if d == p {
			pkg = mu.module.Name

			break
		}
		p = d
	}

	return normalisePkgPath(pkg)
}

func normalisePkgPath(pkg string) string {
	sep := fmt.Sprintf("%c", os.PathSeparator)

	return strings.ReplaceAll(pkg, sep, "/")
}

func (mu *Engine) mutationStatus(pos token.Position) mutator.Status {
	var status mutator.Status

	if mu.codeData.Cov.IsCovered(pos) {
		status = mutator.Runnable
	}

	if !mu.codeData.Diff.IsChanged(pos) {
		status = mutator.Skipped
	}

	return status
}

func (mu *Engine) findNodeMutations(fileName string, set *token.FileSet, file *ast.File, node *Node) {
	// Statement block removal
	var l []ast.Stmt

	switch n := (*node.node).(type) {
	case *ast.BlockStmt:
		l = n.List
	case *ast.CaseClause:
		l = n.Body
	}

	for i, ni := range l {
		if checkRemoveStatement(ni) {
			tm := NewStmtRemover(mu.pkgName(fileName, file.Name.Name), set, file, node, i, ni.Pos())
			tm.SetType(mutator.RemoveStatement)
			tm.SetStatus(mu.mutationStatus(set.Position(tm.Pos())))

			mu.mutants = append(mu.mutants, tm)
		}
	}
}

func checkRemoveStatement(node ast.Stmt) bool {
	if isAridNode(node) {
		return false
	}

	switch n := node.(type) {
	case *ast.AssignStmt:
		return n.Tok != token.DEFINE
	case *ast.IncDecStmt:
		return true
	case *ast.ExprStmt:
		return true
	}

	return false
}

func isAridNode(node ast.Node) bool {
	if node == nil {
		return true
	}

	// Base case
	switch n := node.(type) {
	case *ast.ExprStmt:
		if isLoggerStmt(node) {
			return true
		}

		return isAridNode(n.X)
	case *ast.BlockStmt:
		allChildrenArid := true
		for _, s := range n.List {
			if !isAridNode(s) {
				allChildrenArid = false
				break
			}
		}
		return allChildrenArid
	case *ast.IfStmt:
		return isAridNode(n.Init) && isAridNode(n.Body) && isAridNode(n.Else)
	case *ast.CallExpr:
		return isAridNode(n.Fun)
	case *ast.Ident:
		if n.Obj == nil {
			return true
		}
		if funDecl, ok := (n.Obj.Decl).(*ast.FuncDecl); ok {
			return isAridNode(funDecl)
		}
	case *ast.FuncDecl:
		return isAridNode(n.Body)
	case *ast.CaseClause:
		allChildrenArid := true
		for _, s := range n.Body {
			if !isAridNode(s) {
				allChildrenArid = false
				break
			}
		}
		return allChildrenArid
	}

	return false
}

func isLoggerStmt(es ast.Node) bool {
	firstIdent := ""
	ast.Inspect(es, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			// Since it's a depth first traversal, first identifier should be
			// one of loggerIdentifiers for the statement to be a logger stmt.
			if firstIdent != "" {
				// If we've already found our candidate don't recurse further
				return false
			}
			firstIdent = ident.Name
			return false
		}

		if firstIdent != "" {
			return false
		}
		return true
	})

	_, found := loggerIdentifiers[firstIdent]

	return found
}

func (mu *Engine) executeTests(ctx context.Context) report.Results {
	pool := workerpool.Initialize("mutator")
	pool.Start()

	var mutants []mutator.Mutator
	outCh := make(chan mutator.Mutator)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, mut := range mu.mutants {
			ok := checkDone(ctx)
			if !ok {
				pool.Stop()

				break
			}
			wg.Add(1)
			pool.AppendExecutor(mu.jDealer.NewExecutor(mut, outCh, wg))
		}
	}()

	go func() {
		wg.Wait()
		close(outCh)
	}()

	surfacedMutants := map[string]map[int]bool{}
	comments := []Comment{}

	for m := range outCh {
		mu.logger.Mutant(m)
		mutants = append(mutants, m)

		if m.Status() == mutator.Lived {
			shouldComment := false

			fileMutants, ok := surfacedMutants[m.Position().Filename]
			if !ok {
				fileMutants = map[int]bool{m.Position().Line: true}
				surfacedMutants[m.Position().Filename] = fileMutants
				shouldComment = true
			} else {
				_, ok := fileMutants[m.Position().Line]
				if !ok {
					fileMutants[m.Position().Line] = true
					surfacedMutants[m.Position().Filename] = fileMutants
					shouldComment = true
				}
			}

			if shouldComment {
				comment := Comment{
					Body: getPRComment(m),
					Path: m.Position().Filename,
					Line: m.Position().Line,
					Side: "RIGHT",
				}
				comments = append(comments, comment)
			}
		}
	}

	// Marshal the data into JSON
	jsonData, err := json.MarshalIndent(comments, "", "    ")
	if err != nil {
		log.Fatalf("Error occurred during marshalling. %v", err)
	}

	// Writing the JSON data to a file
	fileName := "comments.json"
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error occurred creating file: %v", err)
	}
	defer file.Close()

	// Write the JSON data to file
	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatalf("Error occurred writing to file: %v", err)
	}
	log.Printf("Data successfully written to %s", fileName)

	return results(mutants)
}

func getPRComment(m mutator.Mutator) string {
	return fmt.Sprintf(
		"[gremlins] Changing the code like shown below does not cause any tests exercising them to fail.\n"+
			"Consider adding tests that fail when the code is mutated.\n\n"+
			"```diff\n%s\n```", m.Diff())
}

func checkDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func results(m []mutator.Mutator) report.Results {
	return report.Results{Mutants: m}
}
