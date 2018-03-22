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
		case "Float":
			return prefix + "double"
		case "Boolean":
			return prefix + "bool"
		case "ID":
			return prefix + "graphql.ID"
		default:
			return "*" + name + "Resolver"
		}
	}
	return prefix + gqlType.String()
}

func getDefaultReturnValue(gqlType ast.Type) string {
	nonNull, ok := gqlType.(*ast.NonNull)
	if !ok {
		return "nil"
	}

	named, ok := nonNull.Type.(*ast.Named)
	if !ok {
		return "nil"
	}

	switch named.Name.Value {
	case "String":
		return "\"\""
	case "ID":
		return "\"\""
	case "Int":
		return "0"
	case "Float":
		return "0"
	case "Boolean":
		return "false"
	}
	return "nil"
}
