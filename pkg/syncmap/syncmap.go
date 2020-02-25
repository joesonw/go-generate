package syncmap

import (
	"fmt"
	"go-generate/pkg/generator"
	"go/ast"
	"runtime"
	"strings"
)

func New(name, pkg, typ string) (g *generator.Generator, err error) {
	gen := &Generator{
		name: name,
	}
	g, err = generator.New(pkg, fmt.Sprintf("%s/src/sync/map.go", runtime.GOROOT()), typ, gen)
	gen.Generator = g
	return g, err
}

type Generator struct {
	*generator.Generator
	name string
}

func (g *Generator) Names() map[string]string {
	return map[string]string{
		"Map":      g.name,
		"entry":    "entry" + strings.Title(g.name),
		"readOnly": "readOnly" + strings.Title(g.name),
		"expunged": "expunged" + strings.Title(g.name),
		"newEntry": "newEntry" + strings.Title(g.name),
	}
}

func (g *Generator) Values() map[string]func(*ast.ValueSpec) {
	return map[string]func(*ast.ValueSpec){
		"expunged": func(v *ast.ValueSpec) { g.ReplaceValue(v) },
	}
}

// Types returns all TypesSpec handlers for AST mutation.
func (g *Generator) Types() map[string]func(*ast.TypeSpec) {
	return map[string]func(*ast.TypeSpec){
		"Map": func(t *ast.TypeSpec) {
			l := t.Type.(*ast.StructType).Fields.List[0]
			l.Type = generator.Expr("sync.Mutex", l.Type.Pos())
			g.ReplaceKey(t.Type)
		},
		"readOnly": func(t *ast.TypeSpec) { g.ReplaceKey(t) },
		"entry":    func(*ast.TypeSpec) {},
	}
}

// Funcs returns all FuncDecl handlers for AST mutation.
func (g *Generator) Funcs() map[string]func(*ast.FuncDecl) {
	nop := func(*ast.FuncDecl) {}
	return map[string]func(*ast.FuncDecl){
		"Load": func(f *ast.FuncDecl) {
			g.ReplaceKey(f.Type.Params)
			g.ReplaceValue(f.Type.Results)
			generator.RenameNil(f.Body, f.Type.Results.List[0].Names[0].Name)
		},
		"load": func(f *ast.FuncDecl) {
			g.ReplaceValue(f)
			generator.RenameNil(f.Body, f.Type.Results.List[0].Names[0].Name)
		},
		"Store": func(f *ast.FuncDecl) {
			g.RenameTuple(f.Type.Params)
		},
		"LoadOrStore": func(f *ast.FuncDecl) {
			g.RenameTuple(f.Type.Params)
			g.ReplaceValue(f.Type.Results)
		},
		"tryLoadOrStore": func(f *ast.FuncDecl) {
			g.ReplaceValue(f)
			generator.RenameNil(f.Body, f.Type.Results.List[0].Names[0].Name)
		},
		"Range": func(f *ast.FuncDecl) {
			g.RenameTuple(f.Type.Params.List[0].Type.(*ast.FuncType).Params)
		},
		"Delete":           func(f *ast.FuncDecl) { g.ReplaceKey(f) },
		"newEntry":         func(f *ast.FuncDecl) { g.ReplaceValue(f) },
		"tryStore":         func(f *ast.FuncDecl) { g.ReplaceValue(f) },
		"dirtyLocked":      func(f *ast.FuncDecl) { g.ReplaceKey(f) },
		"storeLocked":      func(f *ast.FuncDecl) { g.ReplaceValue(f) },
		"delete":           nop,
		"missLocked":       nop,
		"unexpungeLocked":  nop,
		"tryExpungeLocked": nop,
	}
}
