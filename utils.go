package main

import (
	"regexp"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
)

func normalizeName(name string) string {
	re, _ := regexp.Compile("(^|_)([a-z])")
	return re.ReplaceAllStringFunc(name, func(m string) string {
		return strings.ToUpper(strings.Replace(m, "_", "", 1))
	})
}

func genFileName(typeName string) string {
	re, _ := regexp.Compile("[A-Z]")
	name := re.ReplaceAllStringFunc(typeName, func(m string) string {
		return "-" + strings.ToLower(m)
	})
	return strings.Trim(name, "-")
}

func convertGqlType(gqlType ast.Type, nullable bool) string {
	prefix := ""
	if nullable {
		prefix = "*"
	}

	switch gqlType.(type) {
	case *ast.NonNull:
		return convertGqlType(gqlType.(*ast.NonNull).Type, false)
	case *ast.List:
		return prefix + "[]" + convertGqlType(gqlType.(*ast.List).Type, true)
	case *ast.Named:
		name := gqlType.(*ast.Named).Name.Value
		switch name {
		case "String":
			return prefix + "string"
		case "Int":
			return prefix + "int"
		case "ID":
			return "graphql.ID"
		default:
			return "*" + name + "Resolver"
		}
	}
	return prefix + gqlType.String()
}

func findOperationTypes(definitions []ast.Node) []string {
	var types []string
	for _, d := range definitions {
		switch d.(type) {
		case *ast.SchemaDefinition:
			def, _ := d.(*ast.SchemaDefinition)

			for _, operation := range def.OperationTypes {
				types = append(types, operation.Type.Name.Value)
			}

		}
	}
	return types
}
