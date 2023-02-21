package go2cpp

import (
	"fmt"
	"go/ast"
	"log"
	"strings"
)

func parseStmt(fullText []byte, v ast.Stmt) (string, error) {
	switch stmt := v.(type) {
	case *ast.ReturnStmt:
		if len(stmt.Results) == 0 {
			return "return", nil
		}
		str, e := parseExpr(fullText, stmt.Results[0])
		if e != nil {
			log.Println(e)
			return "", e
		}

		return "return " + str + "", nil

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
			buf.WriteString(left + " = " + right + "")
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
		return left + stmt.Tok.String() + "", nil
	case *ast.IfStmt:
		return parseIfStmt(fullText, stmt)
	case *ast.ForStmt:
		return parseForStmt(fullText, stmt)
	default:
		return "", fmt.Errorf("unsupported statement: %s", stringifyNode(fullText, v))
	}
}

func parseBlockStmt(fullText []byte, blockStmt *ast.BlockStmt) (string, error) {
	buf := new(strings.Builder)
	buf.WriteString("{\n")
	for _, stmt := range blockStmt.List {
		s, e := parseStmt(fullText, stmt)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(s + ";\n")
	}
	buf.WriteString("}\n")
	return buf.String(), nil
}

func parseIfStmt(fullText []byte, v *ast.IfStmt) (string, error) {
	buf := new(strings.Builder)
	if v.Init != nil {
		s, e := parseStmt(fullText, v.Init)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(s + ";\n")
	}

	buf.WriteString("if(")
	str, e := parseExpr(fullText, v.Cond)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(str + ")")

	str, e = parseBlockStmt(fullText, v.Body)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(str)
	return buf.String(), nil
}

func parseForStmt(fullText []byte, v *ast.ForStmt) (string, error) {
	buf := new(strings.Builder)
	if v.Init != nil {
		s, e := parseStmt(fullText, v.Init)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(s + ";\n")
	}

	buf.WriteString("for(;")
	str, e := parseExpr(fullText, v.Cond)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(str + ";")

	post, e := parseStmt(fullText, v.Post)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(post + ")\n")

	//block
	block, e := parseBlockStmt(fullText, v.Body)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(block)

	return buf.String(), nil
}
