package singleflight

import (
	"bytes"
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/joesonw/go-generate/pkg/generator"
	"golang.org/x/tools/go/ast/astutil"
)

type versionSlice []*version.Version

var _ sort.Interface = versionSlice([]*version.Version{})

func (s versionSlice) Len() int {
	return len(s)
}

func (s versionSlice) Less(i, j int) bool {
	return !s[i].LessThan(s[j])
}

func (s versionSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func New(name, pkg, typ, ver string) (g *generator.Generator, err error) {
	golangXPath := path.Join(build.Default.GOPATH, "pkg/mod/golang.org/x")
	if _, err := os.Stat(golangXPath); os.IsNotExist(err) {
		generator.Check(err, "please \"go get golang.org/x/sync/singleflight\" first")
	}

	libraryPath := ""
	if ver != "" {
		libraryPath = path.Join(golangXPath, "sync@"+ver)
	} else {

		var syncVersions versionSlice
		xFiles, err := ioutil.ReadDir(golangXPath)
		generator.Check(err, "find versions of golang.org/x/sync/singleflight")
		for _, file := range xFiles {
			name := file.Name()
			if strings.HasPrefix(name, "sync@") {
				v := name[5:]
				vv, _ := version.NewVersion(v)
				syncVersions = append(syncVersions, vv)
			}
		}
		if len(syncVersions) == 0 {
			panic(errors.New("please \"go get golang.org/x/sync/singleflight\" first"))
		}
		sort.Sort(syncVersions)
		libraryPath = path.Join(golangXPath, "sync@v"+syncVersions[0].String())
	}
	println("using singleflight package from: " + libraryPath)

	gen := &Generator{
		name: name,
	}
	g, err = generator.New(pkg, libraryPath+"/singleflight/singleflight.go", gen)

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
	return map[string]func(*ast.ValueSpec){}
}

// Types returns all TypesSpec handlers for AST mutation.
func (g *Generator) Types() map[string]func(*ast.TypeSpec) {
	return map[string]func(*ast.TypeSpec){
		"call": func(n *ast.TypeSpec) {
			generator.ReplaceIface(n, g.value)
		},
		"Group": func(n *ast.TypeSpec) {
			s := n.Type.(*ast.StructType)
			for _, f := range s.Fields.List {
				if m, ok := f.Type.(*ast.MapType); ok {
					astutil.Apply(m, func(c *astutil.Cursor) bool {
						if c.Node().Pos() == m.Key.Pos() {
							c.Replace(generator.Expr(g.key, m.Key.Pos()))
						}
						return true
					}, nil)
				}
			}
		},
		"Result": func(n *ast.TypeSpec) {
			generator.ReplaceIface(n, g.value)
		},
	}
}

// Funcs returns all FuncDecl handlers for AST mutation.
func (g *Generator) Funcs() map[string]func(*ast.FuncDecl) {
	return map[string]func(*ast.FuncDecl){
		"Do": func(f *ast.FuncDecl) {
			generator.ReplaceIface(f.Type.Results, g.value)
			g.replaceKey(f.Type.Params.List[0])
			g.replaceFunction(f.Type.Params.List[1])
			g.replaceMake(f)
		},
		"DoChan": func(f *ast.FuncDecl) {
			g.replaceKey(f.Type.Params.List[0])
			g.replaceFunction(f.Type.Params.List[1])
			g.replaceMake(f)
		},
		"doCall": func(f *ast.FuncDecl) {
			g.replaceKey(f.Type.Params.List[1])
			g.replaceFunction(f.Type.Params.List[2])
		},
		"Forget": func(f *ast.FuncDecl) {
			g.replaceKey(f.Type.Params.List[0])
		},
	}
}

func (g *Generator) Mutate() error {
	g.AddImport("golang.org/x/sync/singleflight")
	g.Rename(map[string]string{
		"Group":  g.name,
		"call":   "call" + strings.Title(g.name),
		"Result": "Result" + strings.Title(g.name),
	})
	return nil
}

func (g *Generator) replaceKey(f *ast.Field) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		n := c.Node()
		if it, ok := n.(*ast.Ident); ok && it.Name == "string" {
			c.Replace(generator.Expr(g.key, it.Pos()))
		}
		return true
	}, nil)
}

func (g *Generator) replaceFunction(f *ast.Field) {
	generator.ReplaceIface(f.Type.(*ast.FuncType).Results.List[0], g.value)
}

func (g *Generator) replaceMake(f *ast.FuncDecl) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		n := c.Node()
		if m, ok := n.(*ast.MapType); ok {
			astutil.Apply(n, func(c *astutil.Cursor) bool {
				if c.Node().Pos() == m.Key.Pos() {
					c.Replace(generator.Expr(g.key, m.Key.Pos()))
				}
				return true
			}, nil)
		}
		return true
	}, nil)
}
