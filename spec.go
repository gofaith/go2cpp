package go2cpp

import (
	"fmt"
	"go/ast"
	"log"
	"strings"
)

func parseSpec(fullText []byte, spec ast.Spec, variables map[string]struct{}) (string, error) {
	switch v := spec.(type) {
	case *ast.ValueSpec:
		buf := new(strings.Builder)
		var e error
		var t = "auto"
		if v.Type != nil {
			t, e = parseFieldType(fullText, v.Type)
			if e != nil {
				log.Println(e)
				return "", e
			}
		}

		for i, name := range v.Names {
			if _, ok := variables[name.Name]; ok {
				return "", fmt.Errorf("variable already declared: %s", name.Name)
			}
			variables[name.Name] = struct{}{}

			buf.WriteString(t + " " + name.Name)
			if len(v.Values) > 0 {
				value, e := parseExpr(fullText, v.Values[i])
				if e != nil {
					log.Println(e)
					return "", e
				}

				buf.WriteString(" = " + value)
			}
			if i < len(v.Names)-1 {
				buf.WriteString(";\n")
			}
		}
		return buf.String(), nil

	default:
		return "", fmt.Errorf("unsupported spec: %s", stringifyNode(fullText, spec))
	}
}
