package go2cpp

import (
	"go/ast"
)

func stringifyFieldList(fullText []byte, l *ast.FieldList) string {
	return string(fullText[l.Opening-1 : l.Closing-1])
}

func stringifyNode(fullText []byte, l ast.Node) string {
	return string(fullText[l.Pos()-1 : l.End()-1])
}
