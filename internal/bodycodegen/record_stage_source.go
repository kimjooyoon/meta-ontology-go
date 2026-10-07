package bodycodegen

import (
	"context"
	"fmt"
)

// RecordAssemblyStageSource projects a historical stage after the enclosing
// generation has been fully replayed. It binds that stage to a composition's
// reconstructed dependency checkpoint; it does not verify scores by itself.
func RecordAssemblyStageSource(ctx context.Context, filename string, source []byte, prior Result,
	stage int, called bool) ([]byte, error) {
	r := prior.Report.RecordAssembly
	if r == nil || stage < 0 || stage >= len(r.ControlHistory) {
		return nil, fmt.Errorf("record stage is missing")
	}
	inputDigest := r.OriginalSourceSHA256
	for _, checkpoint := range r.ControlHistory[:stage+1] {
		if checkpoint.Source != "" {
			inputDigest = checkpoint.SourceSHA256
		}
	}
	if digest(source) != inputDigest {
		return nil, fmt.Errorf("record stage source differs from constructed dependencies")
	}
	p, err := prepareRecordAssembly(ctx, filename, source, prior.Report.Activity)
	if err != nil {
		return nil, err
	}
	mask := r.ControlHistory[stage].SelectedMask
	body, err := p.candidateBody(mask)
	if err != nil {
		return nil, err
	}
	choices := make(map[string]string, len(p.choices))
	for i, choice := range p.choices {
		choices[choice.ID] = "value_first"
		if mask&(1<<i) != 0 {
			choices[choice.ID] = "value_second"
		}
	}
	completed, err := selectedTypedPathSource(source, p.original.activity, body, choices, true)
	if err != nil || !called {
		return completed, err
	}
	fixed, err := fixedCalledSource(filename, string(completed), prior.Report.Activity)
	return []byte(fixed), err
}
