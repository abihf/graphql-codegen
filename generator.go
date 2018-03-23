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
	operationTypeNames := make(map[string]bool)
	interfaceImplementors := make(map[string][]string)
	var operationTypes []*ast.ObjectDefinition

	for _, d := range definitions {
		switch d.(type) {
		case *ast.SchemaDefinition:
			def, _ := d.(*ast.SchemaDefinition)
			for _, operation := range def.OperationTypes {
				operationTypeNames[operation.Type.Name.Value] = true
			}
		case *ast.ObjectDefinition:
			def, _ := d.(*ast.ObjectDefinition)
			for _, iface := range def.Interfaces {
				ifaceName := iface.Name.Value
				interfaceImplementors[ifaceName] = append(interfaceImplementors[ifaceName], def.Name.Value)
			}
		}
	}

	for _, d := range definitions {
		switch d.(type) {
		case *ast.ObjectDefinition:
			def, _ := d.(*ast.ObjectDefinition)
			name := def.Name.Value
			if _, ok := operationTypeNames[name]; ok {
				operationTypes = append(operationTypes, def)
			} else {
				defs := []*ast.ObjectDefinition{def}
				g.generateTypeResolver(genFileName(name), name+"Resolver", defs)
			}

		case *ast.InterfaceDefinition:
			def, _ := d.(*ast.InterfaceDefinition)
			name := def.Name.Value
			g.generateInterfaceResolver(genFileName(name), def, interfaceImplementors[name])

		case *ast.EnumDefinition:
			def, _ := d.(*ast.EnumDefinition)
			name := def.Name.Value
			g.generateEnumResolver(genFileName(name), def)

		}
	}

	g.generateTypeResolver("resolver", "Resolver", operationTypes)
}

func (g *generator) createFile(fileName string) (*os.File, error) {
	fullPath := path.Join(g.dir, fileName+".go")
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}

	header := fmt.Sprintf("package %s\n", g.packageName)
	f.Write([]byte(header))

	const importStatements = `
import (
	"github.com/graph-gophers/graphql-go"
	"golang.org/x/net/context"
)

`
	f.Write([]byte(importStatements))

	return f, nil
}

func (g *generator) generateTypeResolver(fileName, resolverName string, defs []*ast.ObjectDefinition) error {
	f, err := g.createFile(fileName)
	defer f.Close()
	if err != nil {
		return err
	}

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

func (g *generator) generateInterfaceResolver(fileName string, def *ast.InterfaceDefinition, implementors []string) error {
	f, err := g.createFile(fileName)
	defer f.Close()
	if err != nil {
		return err
	}

	ifaceName := def.Name.Value
	f.Write([]byte(fmt.Sprintf("type %s interface {\n", ifaceName)))

	firstField := true
	for _, field := range def.Fields {
		funcName := normalizeName(field.Name.Value)
		fieldType := convertGqlType(field.Type, true)
		var args string
		if len(field.Arguments) > 0 {
			args = ", " + g.generateArgumentStruct(field.Arguments)
		}
		var prefix string
		if firstField {
			firstField = false
		} else {
			prefix = "\n"
		}
		line := prefix + fmt.Sprintf("  // %[2]s resolves %[3]s from %[1]s\n  %[2]s(context.Context%[5]s) (%[4]s, error)\n",
			ifaceName, funcName, field.Name.Value, fieldType, args)
		f.Write([]byte(line))
	}
	f.Write([]byte("}\n\n"))

	resolverName := ifaceName + "Resolver"
	ifaceAbbr := strings.ToLower(string(ifaceName[0]))
	resolverDeclaration := fmt.Sprintf("type %s struct {\n  %s\n}\n", resolverName, ifaceName)
	f.Write([]byte(resolverDeclaration))

	for _, implementor := range implementors {
		funcDeclration := fmt.Sprintf(`
// To%[4]s convert %[2]s to %[4]s
func (%[3]s *%[1]s) To%[4]s() (*%[4]sResolver, bool) {
	res, ok := %[3]s.%[2]s.(*%[4]sResolver)
	return res, ok
}
`, resolverName, ifaceName, ifaceAbbr, implementor)
		f.Write([]byte(funcDeclration))
	}

	return nil
}

func (g *generator) generateEnumResolver(fileName string, def *ast.EnumDefinition) error {
	f, err := g.createFile(fileName)
	defer f.Close()
	if err != nil {
		return err
	}

	name := def.Name.Value
	resolverName := name + "Resolver"
	fmt.Fprintf(f, "type %s string\n\nconst (\n", resolverName)

	for _, val := range def.Values {
		constName := name + normalizeName(strings.ToLower(val.Name.Value))
		comment := "= " + val.Name.Value
		if val.Description != nil {
			comment = val.Description.Value
		}
		fmt.Fprintf(f, "  // %s %s\n", constName, comment)
		fmt.Fprintf(f, "  %s %s = \"%s\"\n\n", constName, resolverName, val.Name.Value)
	}

	fmt.Fprintln(f, ")")

	return nil
}
