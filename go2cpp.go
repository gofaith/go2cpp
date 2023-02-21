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

	buf.WriteString(")")

	// body
	if d.Body != nil {
		s, e := parseBlockStmt(fullText, d.Body)
		if e != nil {
			log.Println(e)
			return "", e
		}
		buf.WriteString(s)
	}

	return buf.String(), nil
}

func parseDecl(fullText []byte, decl ast.Decl, variables map[string]struct{}) (string, error) {
	switch v := decl.(type) {
	case *ast.GenDecl:
		buf := new(strings.Builder)
		for _, spec := range v.Specs {
			s, e := parseSpec(fullText, spec, variables)
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

func parseIdent(fullText []byte, ident *ast.Ident) (string, error) {
	return ident.Name, nil
}

func parseBasicLit(fullText []byte, v *ast.BasicLit) (string, error) {
	return v.Value, nil
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
