package containerheap

import (
	"fmt"
	"go/ast"
	"runtime"

	"github.com/joesonw/go-generate/pkg/generator"
	"golang.org/x/tools/go/ast/astutil"
)

func New(name, pkg, typ string) (g *generator.Generator, err error) {
	gen := &Generator{
		name: name,
	}
	g, err = generator.New(pkg, fmt.Sprintf("%s/src/container/heap/heap.go", runtime.GOROOT()), gen)

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
		"Interface": func(n *ast.TypeSpec) {
			matched := false
			astutil.Apply(n, func(c *astutil.Cursor) bool {
				n := c.Node()
				if _, ok := n.(*ast.Ident); !matched && ok {
					name := g.typ + "Interface"
					if name[0] == '*' {
						name = name[1:]
					}
					c.Replace(generator.Expr(name, n.Pos()))
					matched = true
				}
				if f, ok := n.(*ast.FuncType); ok {
					if len(f.Params.List) > 0 {
						generator.ReplaceIface(f.Params.List[0], g.typ)
					} else {
						generator.ReplaceIface(f.Results.List[0], g.typ)
					}
				}
				return true
			}, nil)
		},
	}
}

func (g *Generator) Funcs() map[string]func(*ast.FuncDecl) {
	return map[string]func(*ast.FuncDecl){
		"Push": func(f *ast.FuncDecl) {
			g.replaceFunctionParams(f, 1)
		},
		"Pop": func(f *ast.FuncDecl) {
			g.replaceFunctionResult(f)
		},
		"Remove": func(f *ast.FuncDecl) {
			g.replaceFunctionResult(f)
		},
	}
}

func (g *Generator) Mutate() error {
	return nil
}

func (g *Generator) replaceFunctionResult(f *ast.FuncDecl) {
	generator.ReplaceIface(f.Type.Results.List[0], g.typ)
}

func (g *Generator) replaceFunctionParams(f *ast.FuncDecl, index int) {
	generator.ReplaceIface(f.Type.Params.List[index], g.typ)
}
