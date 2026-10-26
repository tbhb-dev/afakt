// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestListenerSourceGuard covers composition and UI source, including inactive build tags.
// References are resolved before calls, so aliases and wrappers can't hide constructors.
// This guard and bound-address recording don't inspect OS sockets.
func TestListenerSourceGuard(t *testing.T) {
	t.Parallel()
	for _, dir := range []string{".", "../lifecycle", "../temporal/ui", "../temporal/embedded"} {
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr == nil && entry.IsDir() {
				for _, finding := range guardListenerDirectory(t, path) {
					t.Error(finding)
				}
			}
			return walkErr
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// guardListenerDirectory checks common source with each build-tagged alternative.
// Unsupported combinations fail type checking and require an explicit source selection.
func guardListenerDirectory(t *testing.T, dir string) []string {
	t.Helper()
	positions := token.NewFileSet()
	common, alternatives, imports := listenerSources(t, positions, dir)
	if len(imports) == 0 {
		return nil
	}
	loader := listenerImporter(t, positions, dir, imports)
	if len(alternatives) == 0 {
		alternatives = append(alternatives, nil)
	}
	var findings []string
	for _, alternative := range alternatives {
		files := slices.Clone(common)
		if alternative != nil {
			files = append(files, alternative)
		}
		info := &types.Info{Uses: make(map[*ast.Ident]types.Object)}
		config := types.Config{Importer: loader}
		if _, err := config.Check("guard/"+filepath.Base(dir), positions, files, info); err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			findings = append(findings, guardListenerFile(positions, file, info)...)
		}
	}
	return findings
}

// listenerSources separates common declarations from build-tagged alternatives.
func listenerSources(t *testing.T, positions *token.FileSet, dir string) ([]*ast.File, []*ast.File, []string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var common, alternatives []*ast.File
	var imports []string
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(positions, path, nil, parser.ParseComments)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, imp := range file.Imports {
			imports = append(imports, constant.StringVal(constant.MakeFromLiteral(imp.Path.Value, token.STRING, 0)))
		}
		tagged := slices.ContainsFunc(file.Comments, func(group *ast.CommentGroup) bool {
			return slices.ContainsFunc(group.List, func(comment *ast.Comment) bool {
				return strings.HasPrefix(comment.Text, "//go:build ")
			})
		})
		if tagged {
			alternatives = append(alternatives, file)
		} else {
			common = append(common, file)
		}
	}
	return common, alternatives, imports
}

// listenerImporter reads compiler export data without adding module dependencies.
func listenerImporter(t *testing.T, positions *token.FileSet, dir string, imports []string) types.Importer {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", "list", "-deps", "-export", "-json", "--")
	// The option terminator prevents source imports from supplying Go flags.
	cmd.Args = append(cmd.Args, imports...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("export listener dependencies: %v: %s", err, &stderr)
	}
	exports := make(map[string]string)
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() {
		var pkg struct {
			ImportPath string `json:"ImportPath"`
			Export     string `json:"Export"`
		}
		if err = decoder.Decode(&pkg); err != nil {
			t.Fatal(err)
		}
		exports[pkg.ImportPath] = pkg.Export
	}
	return importer.ForCompiler(positions, "gc", func(path string) (io.ReadCloser, error) {
		file, openErr := os.Open(exports[path])
		if openErr != nil {
			return nil, fmt.Errorf("open exports for %s: %w", path, openErr)
		}
		return file, nil
	})
}

// guardListenerFile permits only the direct bind call in the shared factory.
func guardListenerFile(positions *token.FileSet, file *ast.File, info *types.Info) []string {
	var findings []string
	report := func(node ast.Node, message string) {
		findings = append(findings, fmt.Sprintf("%s: %s", positions.Position(node.Pos()), message))
	}
	const upstream = "github.com/temporalio/ui-server/v2/"
	allowed := []string{"server/api", "server/config", "server/csrf", "server/route", "ui"}
	for _, imp := range file.Imports {
		pkg := constant.StringVal(constant.MakeFromLiteral(imp.Path.Value, token.STRING, 0))
		if imp.Name != nil && imp.Name.Name == "." {
			report(imp, "dot imports could bypass the listener guard")
		}
		if strings.HasPrefix(pkg, "github.com/temporalio/ui-server/") &&
			!slices.Contains(allowed, strings.TrimPrefix(pkg, upstream)) {
			report(imp, "upstream UI must use only handlers and assets")
		}
	}
	for _, decl := range file.Decls {
		findings = append(findings, guardListenerDecl(positions, decl, info)...)
	}
	return findings
}

// guardListenerDecl checks references while counting direct binds in the shared factory.
func guardListenerDecl(positions *token.FileSet, decl ast.Decl, info *types.Info) []string {
	var findings []string
	report := func(node ast.Node, message string) {
		findings = append(findings, fmt.Sprintf("%s: %s", positions.Position(node.Pos()), message))
	}
	fn, ok := decl.(*ast.FuncDecl)
	factory := ok && filepath.Clean(positions.Position(decl.Pos()).Filename) == "server.go" &&
		fn.Recv == nil && fn.Name.Name == "listenTCP"
	var permitted *ast.Ident
	binds := 0
	ast.Inspect(decl, func(node ast.Node) bool {
		if bind := listenerBind(node, info); factory && bind != nil {
			permitted = bind
			binds++
		}
		id, named := node.(*ast.Ident)
		if !named || id == permitted {
			return true
		}
		obj := info.Uses[id]
		if listenerReference(obj, factory) {
			report(id, obj.String()+" bypasses the injectable listener")
		}
		return true
	})
	if factory && binds != 1 {
		report(decl, fmt.Sprintf("listenTCP has %d bind references, want exactly one", binds))
	}
	return findings
}

// listenerBind resolves a direct call to the network configuration's binding method.
func listenerBind(node ast.Node, info *types.Info) *ast.Ident {
	call, called := node.(*ast.CallExpr)
	if !called {
		return nil
	}
	sel, selected := call.Fun.(*ast.SelectorExpr)
	if !selected {
		return nil
	}
	obj, method := info.Uses[sel.Sel].(*types.Func)
	if method && obj.Pkg() != nil && obj.Pkg().Path() == "net" && obj.Name() == "Listen" &&
		obj.Signature().Recv() != nil &&
		types.TypeString(obj.Signature().Recv().Type(), nil) == "*net.ListenConfig" {
		return sel.Sel
	}
	return nil
}

// listenerReference recognizes constructors and listener-owning startup by package identity.
func listenerReference(obj types.Object, factory bool) bool {
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	switch obj.Pkg().Path() {
	case "net":
		if factory && obj.Name() == "ListenConfig" {
			return false
		}
		return strings.HasPrefix(obj.Name(), "Listen") && obj.Name() != "Listener" || obj.Name() == "FileListener"
	case "crypto/tls":
		return obj.Name() == "Listen" || obj.Name() == "NewListener"
	case "net/http":
		return strings.HasPrefix(obj.Name(), "ListenAndServe")
	case "github.com/labstack/echo/v4":
		return strings.HasPrefix(obj.Name(), "Start")
	}
	return false
}
