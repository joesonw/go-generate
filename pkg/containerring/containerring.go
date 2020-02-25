package containerring

import (
	"fmt"
	"go/ast"
	"runtime"
	"strings"

	"github.com/joesonw/go-generate/pkg/generator"
)

func New(name, pkg, typ string) (g *generator.Generator, err error) {
	gen := &Generator{
		name: name,
	}
	g, err = generator.New(pkg, fmt.Sprintf("%s/src/container/ring/ring.go", runtime.GOROOT()), gen)

	gen.typ = typ
	gen.Generator = g
	return g, err
}

type Generator struct {
	*generator.Generator
	name string
	typ  string
}

func (g *Generator) Values() map[string]func(*ast.ValueSpec) {
	return map[string]func(*ast.ValueSpec){}
}

func (g *Generator) Types() map[string]func(*ast.TypeSpec) {
	return map[string]func(*ast.TypeSpec){
		"Ring": func(n *ast.TypeSpec) {
			generator.ReplaceIface(n, g.typ)
		},
	}
}

func (g *Generator) Funcs() map[string]func(*ast.FuncDecl) {
	return map[string]func(*ast.FuncDecl){
		"Do": func(f *ast.FuncDecl) {
			generator.ReplaceIface(f.Type.Params.List[0].Type.(*ast.FuncType).Params.List[0], g.typ)
		},
	}
}

func (g *Generator) Mutate() error {
	g.Rename(map[string]string{
		"Ring": strings.Title(g.name) + "Ring",
	})
	return nil
}

func (g *Generator) replaceFunctionResult(f *ast.FuncDecl) {
	generator.ReplaceIface(f.Type.Results.List[0], g.typ)
}

func (g *Generator) replaceFunctionParams(f *ast.FuncDecl, index int) {
	generator.ReplaceIface(f.Type.Params.List[index], g.typ)
}
