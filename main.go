package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/graphql-go/graphql/language/parser"
)

func main() {
	cmdLine := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cmdLine.Usage = func() {
		fmt.Fprintf(cmdLine.Output(), "Usage: %s <schema.gql>\nOptions:\n", os.Args[0])
		cmdLine.PrintDefaults()
	}
	dirName := cmdLine.String("dir", "resolver/", "output directory")
	packageName := cmdLine.String("package", "resolver", "golang package name")
	cmdLine.Parse(os.Args[1:])

	if cmdLine.NArg() < 1 {
		cmdLine.Usage()
		panic("")
	}
	fileName := cmdLine.Arg(0)
	schemaContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	doc, err := parser.Parse(parser.ParseParams{
		Source: string(schemaContent),
		Options: parser.ParseOptions{
			NoLocation: false,
			NoSource:   false,
		},
	})
	if err != nil {
		panic(err)
	}
	g := &generator{dir: *dirName, packageName: *packageName}
	g.generateResolvers(doc.Definitions)
}
