package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
)

type generator struct {
	packageName string
	dir         string
}

func (g *generator) generateResolvers(definitions []ast.Node) {

	typeMaps := make(map[string]*ast.ObjectDefinition)
	var operationTypeNames []string
	for _, d := range definitions {
		switch d.(type) {
		case *ast.ObjectDefinition:
			def, _ := d.(*ast.ObjectDefinition)
			typeMaps[def.Name.Value] = def

		case *ast.SchemaDefinition:
			def, _ := d.(*ast.SchemaDefinition)
			for _, operation := range def.OperationTypes {
				operationTypeNames = append(operationTypeNames, operation.Type.Name.Value)
			}
		}
	}

	var operationTypes []*ast.ObjectDefinition
	for _, name := range operationTypeNames {
		operationTypes = append(operationTypes, typeMaps[name])
		delete(typeMaps, name)
	}
	g.generateTypeResolver("resolver", "Resolver", operationTypes)

	for name, def := range typeMaps {
		defs := []*ast.ObjectDefinition{def}
		g.generateTypeResolver(genFileName(name), name+"Resolver", defs)
	}
}

func (g *generator) generateTypeResolver(fileName, resolverName string, defs []*ast.ObjectDefinition) error {
	fullPath := path.Join(g.dir, fileName+".go")
	f, err := os.Create(fullPath)

	if err != nil {
		return err
	}
	defer f.Close()

	header := fmt.Sprintf("package %s\n", g.packageName)
	f.Write([]byte(header))

	const importStatements = `
import (
	"github.com/graph-gophers/graphql-go"
	"golang.org/x/net/context"
)

`
	f.Write([]byte(importStatements))

	structDefinition := fmt.Sprintf("// %[1]s implementation\ntype %[1]s struct {}\n", resolverName)
	f.Write([]byte(structDefinition))

	for _, def := range defs {
		g.generateFieldResolvers(f, resolverName, def)
	}

	return nil
}

func (g *generator) generateFieldResolvers(f *os.File, resolverName string, def *ast.ObjectDefinition) error {
	abbr := strings.ToLower(string(resolverName[0]))
	for _, field := range def.Fields {
		functionName := normalizeName(field.Name.Value)
		returnType := convertGqlType(field.Type, true)
		var args string
		if len(field.Arguments) > 0 {
			args = ", args " + g.generateArgumentStruct(field.Arguments)
		}
		f.Write([]byte(fmt.Sprintf("\n// %s resolves %s from %s", functionName, field.Name.Value, def.Name.Value)))
		f.Write([]byte(fmt.Sprintf(`
func (%s *%s) %s(ctx context.Context%s) (%s, error) {
	// impl
	return %s, errors.New("Not Implemented")
}
`, abbr, resolverName, functionName, args, returnType, getDefaultReturnValue(field.Type))))
	}
	return nil
}

func (g *generator) generateArgumentStruct(args []*ast.InputValueDefinition) string {
	result := "struct{\n"
	for _, arg := range args {
		argName := normalizeName(arg.Name.Value)
		argType := convertGqlType(arg.Type, true)
		var description string
		if arg.Description != nil {
			description = " // " + arg.Description.Value
		}
		result += fmt.Sprintf("  %s %s%s\n", argName, argType, description)
	}
	return result + "}"
}
