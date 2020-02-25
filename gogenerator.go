package main

import (
	"flag"
	"go-generate/pkg/generator"
	"go-generate/pkg/syncmap"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	pOut = flag.String("out", "", "")
	pName = flag.String("name", "", "")
	pGenerator = flag.String("generator", "","")
)

func main() {
	var g *generator.Generator
	var err error
	goPackage := os.Getenv("GOPACKAGE")

	flag.Parse()
	name := strings.TrimSpace(*pName)
	program := strings.TrimSpace(*pGenerator)

	out := strings.TrimSpace(*pOut)
	if out == "" {
		out = strings.ToLower(name) + "_gen.go"
	}

	cwd, err := os.Getwd()
	die(err)
	out = path.Join(cwd,out)

	expr := os.Args[len(os.Args) - 1]

	if program == "sync/map" {
		g, err = syncmap.New(name, goPackage, expr)
		die(err)
	} else {
		panic(program + " does not exist")
	}

	die(g.Mutate())

	b, err := g.Generate()
	die(err)

	die(ioutil.WriteFile(out, b, 0644))
	die(err)

	die(exec.Command("goimports", "-w", out).Run())
}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
