package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/joesonw/go-generate/pkg/containerheap"
	"github.com/joesonw/go-generate/pkg/containerlist"
	"github.com/joesonw/go-generate/pkg/containerring"
	"github.com/joesonw/go-generate/pkg/generator"
	"github.com/joesonw/go-generate/pkg/singleflight"
	"github.com/joesonw/go-generate/pkg/syncmap"
)

var (
	pOut       = flag.String("out", "", "")
	pName      = flag.String("name", "", "")
	pGenerator = flag.String("generator", "", "")
	pVersion   = flag.String("version", "", "")
)

func main() {
	var g *generator.Generator
	var err error
	goPackage := os.Getenv("GOPACKAGE")

	flag.Parse()
	name := strings.TrimSpace(*pName)
	program := strings.TrimSpace(*pGenerator)
	version := strings.TrimSpace(*pVersion)

	out := strings.TrimSpace(*pOut)
	if out == "" {
		out = strings.ToLower(name) + "_gen.go"
	}

	cwd, err := os.Getwd()
	die(err)
	out = path.Join(cwd, out)

	expr := os.Args[len(os.Args)-1]

	if program == "sync/map" {
		g, err = syncmap.New(name, goPackage, expr)
		die(err)
	} else if program == "container/list" {
		g, err = containerlist.New(name, goPackage, expr)
		die(err)
	} else if program == "container/ring" {
		g, err = containerring.New(name, goPackage, expr)
		die(err)
	} else if program == "container/heap" {
		g, err = containerheap.New(name, goPackage, expr)
		die(err)
	} else if program == "singleflight" {
		g, err = singleflight.New(name, goPackage, expr, version)
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
