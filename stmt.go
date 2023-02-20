package go2cpp

import (
	"fmt"
	"go/ast"
	"log"
	"strings"
)

func parseBlockStmt(fullText []byte, blockStmt *ast.BlockStmt) (string, error) {
	buf := new(strings.Builder)
	buf.WriteString("{\n")
	for _, stmt := range blockStmt.List {
		s, e := parseStmt(fullText, stmt)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(s)
	}
	buf.WriteString("}\n")
	return buf.String(), nil
}

func parseStmt(fullText []byte, v ast.Stmt) (string, error) {
	switch stmt := v.(type) {
	case *ast.ReturnStmt:
		if len(stmt.Results) == 0 {
			return "return;\n", nil
		}
		str, e := parseExpr(fullText, stmt.Results[0])
		if e != nil {
			log.Println(e)
			return "", e
		}

		return "return " + str + ";\n", nil

	case *ast.AssignStmt:
		if stmt.Tok.String() == ":=" {
			return "", fmt.Errorf("unsupported statement: %s", stringifyNode(fullText, v))
		}

		buf := new(strings.Builder)

		for i, hs := range stmt.Lhs {
			left, e := parseExpr(fullText, hs)
			if e != nil {
				log.Println(e)
				return "", e
			}
			right, e := parseExpr(fullText, stmt.Rhs[i])
			if e != nil {
				log.Println(e)
				return "", e
			}
			buf.WriteString(left + " = " + right + ";\n")
		}

		return buf.String(), nil

	case *ast.DeclStmt:
		return parseDecl(fullText, stmt.Decl)
	case *ast.IncDecStmt:
		left, e := parseExpr(fullText, stmt.X)
		if e != nil {
			log.Println(e)
			return "", e
		}
		return left + stmt.Tok.String() + ";\n", nil
	case *ast.IfStmt:
		return parseIfStmt(fullText, stmt)
	default:
		return "", fmt.Errorf("unsupported statement: %s", stringifyNode(fullText, v))
	}
}

func parseIfStmt(fullText []byte, v *ast.IfStmt) (string, error) {
	if v.Init != nil {
		return "", fmt.Errorf("unsupported if statement: %s", stringifyNode(fullText, v))
	}

	buf := new(strings.Builder)
	buf.WriteString("if(")
	str, e := parseExpr(fullText, v.Cond)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(str)
	buf.WriteString(")")

	str, e = parseBlockStmt(fullText, v.Body)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(str)
	return buf.String(), nil
}
