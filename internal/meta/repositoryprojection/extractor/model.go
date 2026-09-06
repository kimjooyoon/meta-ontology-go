package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"sort"
)

type Failure struct {
	Stage, Step, Reason, UnknownClass, NextOperation string
	BlockedBy                                        []string
	Diagnostics                                      []string
}

func (e Failure) Error() string {
	return fmt.Sprintf("%s/%s/%s unknown_class=%s next=%s blocked_by=%v", e.Stage, e.Step, e.Reason, e.UnknownClass, e.NextOperation, e.BlockedBy)
}

func fail(stage, step, reason, class, next string, blocked []string) error {
	return Failure{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: append([]string{}, blocked...)}
}

func failWithDiagnostics(stage, step, reason, class, next string, diagnostics []string) error {
	return Failure{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: []string{}, Diagnostics: append([]string{}, diagnostics...)}
}

type declaration struct {
	node              ast.Decl
	start, end, order int
	identity          string
}

type importSpec struct {
	group      *ast.GenDecl
	spec       *ast.ImportSpec
	path, name string
}

type rendered struct{ source, helper []byte }
type edit struct {
	start, end  int
	replacement []byte
}

func applyEdits(source []byte, edits []edit) ([]byte, error) {
	sort.SliceStable(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	var out bytes.Buffer
	last := 0
	for _, item := range edits {
		if item.start < last || item.start < 0 || item.end < item.start || item.end > len(source) {
			return nil, fail("rewrite-source", "apply-edits", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", nil)
		}
		out.Write(source[last:item.start])
		out.Write(item.replacement)
		last = item.end
	}
	out.Write(source[last:])
	return out.Bytes(), nil
}
