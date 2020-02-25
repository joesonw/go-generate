package syncmap

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"runtime"
	"strings"

	"github.com/joesonw/go-generate/pkg/generator"
)

func New(name, pkg, typ string) (g *generator.Generator, err error) {
	gen := &Generator{
		name: name,
	}
	g, err = generator.New(pkg, fmt.Sprintf("%s/src/sync/map.go", runtime.GOROOT()), gen)

	exp, err := parser.ParseExpr(typ)
	generator.Check(err, "parse expr: %s", typ)
	m, ok := exp.(*ast.MapType)
	generator.Expect(ok, "invalid argument. expected map[T1]T2")
	b := bytes.NewBuffer(nil)
	err = g.FormatNode(b, m.Key)
	generator.Check(err, "format map key")
	gen.key = b.String()
	b.Reset()
	err = g.FormatNode(b, m.Value)
	generator.Check(err, "format map value")
	gen.value = b.String()

	gen.Generator = g
	return g, err
}

type Generator struct {
	*generator.Generator
	name  string
	key   string
	value string
}

func (g *Generator) Values() map[string]func(*ast.ValueSpec) {
	return map[string]func(*ast.ValueSpec){
		"expunged": func(v *ast.ValueSpec) { g.replaceValue(v) },
	}
}

func (g *Generator) Types() map[string]func(*ast.TypeSpec) {
	return map[string]func(*ast.TypeSpec){
		"Map": func(t *ast.TypeSpec) {
			l := t.Type.(*ast.StructType).Fields.List[0]
			l.Type = generator.Expr("sync.Mutex", l.Type.Pos())
			g.replaceKey(t.Type)
		},
		"readOnly": func(t *ast.TypeSpec) { g.replaceKey(t) },
	}
}

func (g *Generator) Funcs() map[string]func(*ast.FuncDecl) {
	return map[string]func(*ast.FuncDecl){
		"Load": func(f *ast.FuncDecl) {
			g.replaceKey(f.Type.Params)
			g.replaceValue(f.Type.Results)
			generator.RenameNil(f.Body, f.Type.Results.List[0].Names[0].Name)
		},
		"load": func(f *ast.FuncDecl) {
			g.replaceValue(f)
			generator.RenameNil(f.Body, f.Type.Results.List[0].Names[0].Name)
		},
		"Store": func(f *ast.FuncDecl) {
			g.renameTuple(f.Type.Params)
		},
		"LoadOrStore": func(f *ast.FuncDecl) {
			g.renameTuple(f.Type.Params)
			g.replaceValue(f.Type.Results)
		},
		"tryLoadOrStore": func(f *ast.FuncDecl) {
			g.replaceValue(f)
			generator.RenameNil(f.Body, f.Type.Results.List[0].Names[0].Name)
		},
		"Range": func(f *ast.FuncDecl) {
			g.renameTuple(f.Type.Params.List[0].Type.(*ast.FuncType).Params)
		},
		"Delete":      func(f *ast.FuncDecl) { g.replaceKey(f) },
		"newEntry":    func(f *ast.FuncDecl) { g.replaceValue(f) },
		"tryStore":    func(f *ast.FuncDecl) { g.replaceValue(f) },
		"dirtyLocked": func(f *ast.FuncDecl) { g.replaceKey(f) },
		"storeLocked": func(f *ast.FuncDecl) { g.replaceValue(f) },
	}
}

func (g *Generator) replaceKey(n ast.Node) { generator.ReplaceIface(n, g.key) }

func (g *Generator) replaceValue(n ast.Node) { generator.ReplaceIface(n, g.value) }

func (g *Generator) renameTuple(l *ast.FieldList) {
	if g.key == g.value {
		g.replaceKey(l.List[0])
		return
	}
	l.List = append(l.List, &ast.Field{
		Names: []*ast.Ident{l.List[0].Names[1]},
		Type:  l.List[0].Type,
	})
	l.List[0].Names = l.List[0].Names[:1]
	g.replaceKey(l.List[0])
	g.replaceValue(l.List[1])
}

func (g *Generator) Mutate() error {
	g.AddImport("sync")
	g.Rename(map[string]string{
		"Map":      g.name,
		"entry":    "entry" + strings.Title(g.name),
		"readOnly": "readOnly" + strings.Title(g.name),
		"expunged": "expunged" + strings.Title(g.name),
		"newEntry": "newEntry" + strings.Title(g.name),
	})
	return nil
}
