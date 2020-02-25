package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"

	"golang.org/x/tools/go/ast/astutil"
)

func ReplaceIface(n ast.Node, s string) {
	astutil.Apply(n, func(c *astutil.Cursor) bool {
		n := c.Node()
		if it, ok := n.(*ast.InterfaceType); ok {
			c.Replace(Expr(s, it.Interface))
		}
		return true
	}, nil)
}

func RenameNil(n ast.Node, name string) {
	astutil.Apply(n, func(c *astutil.Cursor) bool {
		if _, ok := c.Parent().(*ast.ReturnStmt); ok {
			if i, ok := c.Node().(*ast.Ident); ok && i.Name == new(types.Nil).String() {
				i.Name = name
			}
		}
		return true
	}, nil)
}

func Rename(f *ast.File, oldnew map[string]string) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.Ident:
			if name, ok := oldnew[n.Name]; ok {
				n.Name = name
				n.Obj.Name = name
			}
		case *ast.FuncDecl:
			if name, ok := oldnew[n.Name.Name]; ok {
				n.Name.Name = name
			}
		}
		return true
	}, nil)
}

func Expr(s string, pos token.Pos) ast.Expr {
	exp, err := parser.ParseExpr(s)
	Check(err, "parse expr: %q", s)
	SetPos(exp, pos)
	return exp
}

func SetPos(n ast.Node, p token.Pos) {
	if reflect.ValueOf(n).IsNil() {
		return
	}
	switch n := n.(type) {
	case *ast.Ident:
		n.NamePos = p
	case *ast.MapType:
		n.Map = p
		SetPos(n.Key, p)
		SetPos(n.Value, p)
	case *ast.FieldList:
		n.Closing = p
		n.Opening = p
		if len(n.List) > 0 {
			SetPos(n.List[0], p)
		}
	case *ast.Field:
		SetPos(n.Type, p)
		if len(n.Names) > 0 {
			SetPos(n.Names[0], p)
		}
	case *ast.FuncType:
		n.Func = p
		SetPos(n.Params, p)
		SetPos(n.Results, p)
	case *ast.ArrayType:
		n.Lbrack = p
		SetPos(n.Elt, p)
	case *ast.StructType:
		n.Struct = p
		SetPos(n.Fields, p)
	case *ast.SelectorExpr:
		SetPos(n.X, p)
		n.Sel.NamePos = p
	case *ast.InterfaceType:
		n.Interface = p
		SetPos(n.Methods, p)
	case *ast.StarExpr:
		n.Star = p
		SetPos(n.X, p)
	case *ast.ChanType:
		SetPos(n.Value, p)
	case *ast.ParenExpr:
		SetPos(n.X, p)
	default:
		panic(fmt.Sprintf("unknown type: %v", n))
	}
}

// Check panics if the error is not nil.
func Check(err error, msg string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		panic(genError{fmt.Sprintf(msg+": %s", args...)})
	}
}

// Expect panic if the condition is false.
func Expect(cond bool, msg string, args ...interface{}) {
	if !cond {
		panic(genError{fmt.Sprintf(msg, args...)})
	}
}

type genError struct {
	msg string
}

func (p genError) Error() string { return fmt.Sprintf("go-generate: %s", p.msg) }

func Catch(err *error) {
	if e := recover(); e != nil {
		gerr, ok := e.(genError)
		if !ok {
			panic(e)
		}
		*err = gerr
	}
}

func FailOnErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err.Error())
		os.Exit(1)
	}
}
