package go2cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

func parseStmt(fullText []byte, v ast.Stmt, variables map[string]struct{}) (string, error) {
	switch stmt := v.(type) {
	case *ast.ReturnStmt:
		if len(stmt.Results) == 0 {
			return "return", nil
		}
		if len(stmt.Results) > 1 {
			// unimplemented, because of try catch block problem
			return "", fmt.Errorf("Multiple return values are not supported: %v", stringifyNode(fullText, v))

			// second := stringifyNode(fullText, stmt.Results[1])
			// if second != "nil" {
			// 	if len(stmt.Results) > 2 || !(strings.HasPrefix(second, "errors.New(") || strings.HasPrefix(second, "fmt.Errorf(")) {
			// 		return "", fmt.Errorf("Multiple return values are not supported: %v", stringifyNode(fullText, v))
			// 	}
			// 	// error
			// 	exception, e := parseExpr(fullText, stmt.Results[1])
			// 	if e != nil {
			// 		log.Println(e)
			// 		return "", e
			// 	}
			// 	return "throw " + exception, nil
			// }
		}

		str, e := parseExpr(fullText, stmt.Results[0])
		if e != nil {
			log.Println(e)
			return "", e
		}

		return "return " + str, nil

	case *ast.AssignStmt:
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

			// a:=1
			if _, ok := hs.(*ast.Ident); ok && stmt.Tok == token.DEFINE {
				if _, ok := variables[left]; !ok {
					buf.WriteString("auto ")
					variables[left] = struct{}{}
				}
			}
			buf.WriteString(left + " = " + right + "")
			if i < len(stmt.Lhs)-1 {
				buf.WriteString(";\n")
			}
		}

		return buf.String(), nil

	case *ast.DeclStmt:
		return parseDecl(fullText, stmt.Decl, variables)
	case *ast.IncDecStmt:
		left, e := parseExpr(fullText, stmt.X)
		if e != nil {
			log.Println(e)
			return "", e
		}
		return left + stmt.Tok.String() + "", nil
	case *ast.IfStmt:
		return parseIfStmt(fullText, stmt, variables)
	case *ast.ForStmt:
		return parseForStmt(fullText, stmt, variables)
	default:
		return "", fmt.Errorf("unsupported statement: %s", stringifyNode(fullText, v))
	}
}

func parseBlockStmt(fullText []byte, blockStmt *ast.BlockStmt) (string, error) {
	buf := new(strings.Builder)
	buf.WriteString("{\n")
	variables := make(map[string]struct{})
	for _, stmt := range blockStmt.List {
		s, e := parseStmt(fullText, stmt, variables)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(s + ";\n")
	}
	buf.WriteString("}\n")
	return buf.String(), nil
}

func parseIfStmt(fullText []byte, v *ast.IfStmt, variables map[string]struct{}) (string, error) {
	buf := new(strings.Builder)
	if v.Init != nil {
		s, e := parseStmt(fullText, v.Init, variables)
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

func parseForStmt(fullText []byte, v *ast.ForStmt, variables map[string]struct{}) (string, error) {
	buf := new(strings.Builder)
	if v.Init != nil {
		s, e := parseStmt(fullText, v.Init, variables)
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

	post, e := parseStmt(fullText, v.Post, variables)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(post + ")")

	//block
	block, e := parseBlockStmt(fullText, v.Body)
	if e != nil {
		log.Println(e)
		return "", e
	}
	buf.WriteString(block)

	return buf.String(), nil
}
