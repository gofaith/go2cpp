package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
)

func parseExpr(fullText []byte, expr ast.Expr) (string, error) {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return parseBasicLit(fullText, v)
	case *ast.Ident:
		return parseIdent(fullText, v)
	case *ast.BinaryExpr:
		return parseBinaryExpr(fullText, v)
	default:
		return "", fmt.Errorf("unsupported expression: %s", stringifyNode(fullText, expr))
	}
}

func parseBinaryExpr(fullText []byte, binary *ast.BinaryExpr) (string, error) {
	switch binary.Op {
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
		token.AND,            // &
		token.OR,             // |
		token.XOR,            // ^
		token.SHL,            // <<
		token.SHR,            // >>
		token.AND_NOT,        // &^
		token.ADD_ASSIGN,     // +=
		token.SUB_ASSIGN,     // -=
		token.MUL_ASSIGN,     // *=
		token.QUO_ASSIGN,     // /=
		token.REM_ASSIGN,     // %=
		token.AND_ASSIGN,     // &=
		token.OR_ASSIGN,      // |=
		token.XOR_ASSIGN,     // ^=
		token.SHL_ASSIGN,     // <<=
		token.SHR_ASSIGN,     // >>=
		token.AND_NOT_ASSIGN, // &^=
		token.LAND,           // &&
		token.LOR,            // ||
		token.ARROW,          // <-
		token.INC,            // ++
		token.DEC,            // --
		token.EQL,            // ==
		token.LSS,            // <
		token.GTR,            // >
		token.ASSIGN,         // =
		token.NOT,            // !
		token.NEQ,            // !=
		token.LEQ,            // <=
		token.GEQ,            // >=
		token.DEFINE:         // :=
	default:
		return "", fmt.Errorf("unsupported binary expr: %s", stringifyNode(fullText, binary))
	}

	left, e := parseExpr(fullText, binary.X)
	if e != nil {
		log.Println(e)
		return "", e
	}
	right, e := parseExpr(fullText, binary.Y)
	if e != nil {
		log.Println(e)
		return "", e
	}
	return left + " " + binary.Op.String() + " " + right, nil
}
