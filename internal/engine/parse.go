package engine

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// ParseAndTypeCheckFile parses and type-checks the given file, and returns everything interesting about the file.
// If a fatal error is encountered the error return argument is not nil.
func (mu *Engine) parseAndTypeCheckFile(fileName string) (*ast.File, *token.FileSet, *types.Info, error) {
	src, _ := mu.fs.Open(fileName)
	// Parse source
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, src, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	fileAbs, err := filepath.Abs(fileName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not absolute the file path of %q: %v", fileName, err)
	}
	dir := filepath.Dir(fileAbs)

	// Load complete type information for the specified packages,
	// along with type-annotated syntax.
	// Types for dependencies are loaded from export data.
	conf := &packages.Config{Mode: packages.NeedTypesInfo, Dir: dir}
	pkgs, err := packages.Load(conf)
	if err != nil {
		return nil, nil, nil, err // failed to load anything
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, nil, nil, err // some packages contained errors
	}

	// Find the package and package-level object.
	pkg := pkgs[0]

	return f, fset, pkg.TypesInfo, nil
}
