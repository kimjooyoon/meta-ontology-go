package syntaxregistration

import (
	"fmt"
	"go/ast"
	"strings"
)

func (set admissionSet) appendVersion(source *goSource, version int) error {
	if len(set.rows) == 0 || version <= set.rows[len(set.rows)-1].version {
		return fmt.Errorf("denominator admission cannot overwrite historical versions")
	}
	row := headerAdmissionRow(version)
	if set.argument == "digest" {
		row = admissionRow{version, digestName(version), "true"}
	}
	if len(set.helpers) == 0 {
		set.rows = append(set.rows, row)
		set.normalize(source)
		return nil
	}
	last := set.helpers[len(set.helpers)-1]
	selection := last.Body.List[0].(*ast.SwitchStmt)
	if len(selection.Body.List) < admissionChunkRows {
		source.insert(selection.Body.Rbrace, fmt.Sprintf("case %s: return %s\n", row.key, row.result))
		return nil
	}
	index := len(set.helpers)
	source.replace(last.Body.List[1], "return "+set.helperName(index)+"("+set.argument+")")
	source.appendAt(last, set.renderChunk(index, []admissionRow{row}, true))
	return nil
}

func (set admissionSet) normalize(source *goSource) {
	source.replace(set.root, set.rootCall())
	var text strings.Builder
	for start := 0; start < len(set.rows); start += admissionChunkRows {
		end := min(start+admissionChunkRows, len(set.rows))
		text.WriteString(set.renderChunk(start/admissionChunkRows, set.rows[start:end], end == len(set.rows)))
	}
	source.appendAt(set.root, text.String())
}

func (set admissionSet) renderChunk(index int, rows []admissionRow, terminal bool) string {
	var text strings.Builder
	fmt.Fprintf(&text, "\nfunc %s(%s %s) bool {\nswitch %s {\n",
		set.helperName(index), set.argument, set.parameter, set.tag)
	for _, row := range rows {
		fmt.Fprintf(&text, "case %s: return %s\n", row.key, row.result)
	}
	text.WriteString("}\nreturn ")
	if terminal {
		text.WriteString("false")
	} else {
		fmt.Fprintf(&text, "%s(%s)", set.helperName(index+1), set.argument)
	}
	text.WriteString("\n}\n")
	return text.String()
}
