package go2cpp

import (
	"fmt"
	"go/ast"
)

func parseFieldType(fullText []byte, v ast.Expr) (string, error) {
	vs, ok := v.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("Unsupported field type: %s", stringifyNode(fullText, v))
	}
	s := fmt.Sprint(vs)
	switch s {
	case "string":
		return "QString", nil
	case "int8":
		return "signed char", nil
	case "uint8", "byte":
		return "unsigned char", nil
	case "int16":
		return "short int", nil
	case "uint16":
		return "unsigned short int", nil
	case "int32":
		return "long int", nil
	case "uint32":
		return "unsigned long int", nil
	case "int", "int64":
		return "long long int", nil
	case "uint", "uint64":
		return "unsigned long long int", nil
	case "rune":
		return "QChar", nil
	case "bool":
		return "bool", nil
	default:
		return "", fmt.Errorf("Unsupported field type: %s", stringifyNode(fullText, v))
	}
}
