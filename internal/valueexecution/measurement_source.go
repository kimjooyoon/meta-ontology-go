package valueexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

// Counterfactuals change the selected activity's literal only. A stale or
// ambiguous source yields no candidate, so it cannot report sensitivity.
func measurementSources(filename string, source []byte, program Program) ([]byte, []byte) {
	if digestBytes(source) != program.SourceDigest {
		return nil, nil
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return nil, nil
	}
	var target *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != program.Activity {
			continue
		}
		if target != nil {
			return nil, nil
		}
		target = activity
	}
	if target == nil || !target.ValueProgramPresent || target.ValueProgram != program.Text {
		return nil, nil
	}
	baseline, err := syntax.ReplaceStringLiteral(source, target.ValueProgramSpan, "")
	if err != nil {
		return nil, nil
	}
	changed, err := syntax.ReplaceStringLiteral(source, target.ValueProgramSpan, "int.add:2")
	if err != nil {
		return nil, nil
	}
	return baseline, changed
}
