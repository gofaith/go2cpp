package go2cpp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func ConvertFile(dst, src string) error {
	fset := token.NewFileSet()
	fullText, e := ioutil.ReadFile(src)
	if e != nil {
		log.Println(e)
		return e
	}

	file, e := parser.ParseFile(fset, src, nil, parser.ParseComments)
	if e != nil {
		return e
	}
	fo, e := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if e != nil {
		return e
	}
	defer fo.Close()

	for _, decl := range file.Decls {
		if d, ok := decl.(*ast.FuncDecl); ok {
			s, e := parseFuncDecl(fullText, d)
			if e != nil {
				log.Println(e)
				return e
			}

			fo.Write([]byte(s))
		}
	}
	return nil
}

func parseFuncDecl(fullText []byte, d *ast.FuncDecl) (string, error) {
	buf := new(strings.Builder)

	// return values
	if d.Type != nil && d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		if len(d.Type.Results.List) > 1 {
			return "", fmt.Errorf("multiple return values are not supported in C++: (%s)", stringifyFieldList(fullText, d.Type.Results))
		}
		r, e := parseFieldType(fullText, d.Type.Results.List[0].Type)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(r + " ")
	} else {
		buf.WriteString("void ")
	}

	if d.Name != nil {
		buf.WriteString(d.Name.Name)
	}
	buf.WriteString("(")

	// args
	if d.Type != nil && d.Type.Params != nil {
		ss := []string{}
		for _, arg := range d.Type.Params.List {
			s, e := parseField(fullText, arg)
			if e != nil {
				return "", e
			}
			ss = append(ss, s)
		}
		buf.WriteString(strings.Join(ss, ", "))
	}

	buf.WriteString(") {\n")

	// body
	if d.Body != nil {
		for _, stmt := range d.Body.List {
			s, e := parseStmt(fullText, stmt)
			if e != nil {
				log.Println(e)
				return "", e
			}
			buf.WriteString(s)
		}
	}
	buf.WriteString("}")
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
	default:
		return "", fmt.Errorf("unsupported statement: %s", stringifyNode(fullText, v))
	}
}

func parseDecl(fullText []byte, decl ast.Decl) (string, error) {
	switch v := decl.(type) {
	case *ast.GenDecl:
		buf := new(strings.Builder)
		for _, spec := range v.Specs {
			s, e := parseSpec(fullText, spec)
			if e != nil {
				log.Println(e)
				return "", e
			}
			buf.WriteString(s)
		}
		return buf.String(), nil

	default:
		return "", fmt.Errorf("unsupported declaration: %s", stringifyNode(fullText, decl))
	}
}

func parseSpec(fullText []byte, spec ast.Spec) (string, error) {
	switch v := spec.(type) {
	case *ast.ValueSpec:
		if v.Type == nil {
			return "", fmt.Errorf("unsupported spec: %s", stringifyNode(fullText, spec))
		}
		buf := new(strings.Builder)
		t, e := parseFieldType(fullText, v.Type)
		if e != nil {
			log.Println(e)
			return "", e
		}

		for i, name := range v.Names {
			buf.WriteString(t + " " + name.Name)
			if len(v.Values) > 0 {
				value, e := parseExpr(fullText, v.Values[i])
				if e != nil {
					log.Println(e)
					return "", e
				}

				buf.WriteString(" = " + value)
			}
			buf.WriteString(";\n")
		}
		return buf.String(), nil

	default:
		return "", fmt.Errorf("unsupported spec: %s", stringifyNode(fullText, spec))
	}
}


func parseIdent(fullText []byte, ident *ast.Ident) (string, error) {
	return ident.Name, nil
}

func parseBasicLit(fullText []byte, v *ast.BasicLit) (string, error) {
	return v.Value, nil
}

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
	default:
		return "", fmt.Errorf("Unsupported field type: %s", stringifyNode(fullText, v))
	}
}

func parseField(fullText []byte, v *ast.Field) (string, error) {
	t, e := parseFieldType(fullText, v.Type)
	if e != nil {
		log.Println(e)
		return "", e
	}
	//name
	if len(v.Names) == 0 {
		return "", fmt.Errorf("field value is empty: %v", v)
	}
	return t + " " + v.Names[0].Name, nil
}
