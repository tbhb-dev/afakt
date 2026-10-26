// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestListenerSourceGuard covers composition and UI source, including inactive build tags.
// Upstream UI imports are limited to assets and handler helpers: mux.Handle accepts
// ui.Service as http.Handler, and Service.ServeHTTP delegates to Echo.ServeHTTP.
// Neither the upstream server constructor nor Echo's listener startup is used.
// This guard and bound-address recording use no OS-level socket enumeration.
func TestListenerSourceGuard(t *testing.T) {
	t.Parallel()
	positions := token.NewFileSet()
	for _, dir := range []string{".", "../lifecycle", "../temporal/ui", "../temporal/embedded"} {
		err := filepath.WalkDir(dir, func(path string, _ fs.DirEntry, walkErr error) error {
			if walkErr == nil && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				guardListenerPath(t, positions, path)
			}
			return walkErr
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func guardListenerPath(t *testing.T, positions *token.FileSet, path string) {
	t.Helper()
	file, err := parser.ParseFile(positions, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	networkPackage := guardListenerImports(t, positions, file)
	for _, decl := range file.Decls {
		guardListenerDecl(t, positions, path, networkPackage, decl)
	}
}

func guardListenerImports(t *testing.T, positions *token.FileSet, file *ast.File) string {
	t.Helper()
	const upstream = "github.com/temporalio/ui-server/v2/"
	networkPackage := "net"
	allowed := []string{"server/api", "server/config", "server/csrf", "server/route", "ui"}
	for _, imp := range file.Imports {
		pkg := constant.StringVal(constant.MakeFromLiteral(imp.Path.Value, token.STRING, 0))
		if imp.Name != nil && imp.Name.Name == "." {
			t.Errorf("%s: dot imports could bypass the listener guard", positions.Position(imp.Pos()))
		}
		if pkg == "net" && imp.Name != nil {
			networkPackage = imp.Name.Name
		}
		if strings.HasPrefix(pkg, "github.com/temporalio/ui-server/") &&
			!slices.Contains(allowed, strings.TrimPrefix(pkg, upstream)) {
			t.Errorf("%s: upstream UI must use only handlers and assets", positions.Position(imp.Pos()))
		}
	}
	return networkPackage
}

func guardListenerDecl(t *testing.T, positions *token.FileSet, path, networkPackage string, decl ast.Decl) {
	t.Helper()
	fn, ok := decl.(*ast.FuncDecl)
	factory := ok && path == "server.go" && fn.Recv == nil && fn.Name.Name == "listenTCP"
	binds := 0
	ast.Inspect(decl, func(node ast.Node) bool {
		sel, selected := node.(*ast.SelectorExpr)
		if !selected {
			return true
		}
		// Check references too, so aliasing a forbidden function cannot hide a call.
		if factory && sel.Sel.Name == "ListenConfig" {
			return true
		}
		receiver, named := sel.X.(*ast.Ident)
		if factory && named && receiver.Name == "config" && sel.Sel.Name == "Listen" {
			binds++
			return true
		}
		if listenerReference(sel, networkPackage, path) {
			t.Errorf("%s: %s bypasses the injectable listener", positions.Position(sel.Pos()), sel.Sel.Name)
		}
		return true
	})
	if factory && binds != 1 {
		t.Errorf("listenTCP has %d bind references, want exactly one", binds)
	}
}

func listenerReference(sel *ast.SelectorExpr, networkPackage, path string) bool {
	if strings.Contains(filepath.ToSlash(path), "/ui/") && strings.HasPrefix(sel.Sel.Name, "Start") {
		return true
	}
	if sel.Sel.Name == "Listen" {
		receiver, named := sel.X.(*ast.Ident)
		return named && receiver.Name == networkPackage
	}
	return strings.HasPrefix(sel.Sel.Name, "Listen") && sel.Sel.Name != "Listener"
}
