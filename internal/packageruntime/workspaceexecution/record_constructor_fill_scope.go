package workspaceexecution

import (
	"go/scanner"
	"go/token"
	"sort"
)

// Analyze a fill in its actual source slot before shortening its type name.
// A free alias in a standalone expression may be a local value at that slot.
func (n workspaceRecordNames) rewriteConstructorFill(pkg, hole, expression string,
	scopes []recordConstructorScope) (string, error) {
	if len(scopes) != 0 && scopes[0].body != "" && hole != "" && expression != "" {
		context := constructorFillContext(scopes[0].body, hole, expression)
		if _, err := n.rewriteConstructors(pkg, context, true, scopes...); err != nil {
			return "", err
		}
	}
	return n.rewriteConstructors(pkg, expression, false, scopes...)
}

func constructorFillContext(body, hole, expression string) string {
	file := token.NewFileSet().AddFile("constructor-fill", -1, len(body))
	var scan scanner.Scanner
	scan.Init(file, []byte(body), nil, scanner.ScanComments)
	placeholder := "__GOOO_BODY_HOLE_" + hole + "__"
	var starts []int
	for {
		pos, kind, text := scan.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && text == placeholder {
			starts = append(starts, file.Offset(pos))
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(starts)))
	for _, start := range starts {
		body = body[:start] + expression + body[start+len(placeholder):]
	}
	return body
}
