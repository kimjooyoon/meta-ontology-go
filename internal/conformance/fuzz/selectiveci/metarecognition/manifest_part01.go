package metarecognition

import (
	"fmt"
)

func Run(cases []Case) (Manifest, error) {
	manifest := Manifest{Schema: SchemaVersion, Finding: NoUniqueBenefit, Summary: Summary{ExactOutcomeVector: true, ExactReasonLocalizationVector: true}}
	for _, raw := range cases {
		c := raw.normalized()
		gooo, baseline := Evaluate(c)
		if !gooo.State.Valid() || !gooo.Reason.Valid() || !baseline.State.Valid() || !baseline.Reason.Valid() {
			return Manifest{}, fmt.Errorf("invalid result for %s", c.ID)
		}
		item := ComparisonCase{ID: c.ID, Expected: c.Expected, Gooo: gooo, Baseline: baseline, ExactOutcomeVector: gooo.State == baseline.State && gooo.State == c.Expected.State, ExactReasonVector: gooo.Reason == baseline.Reason && gooo.Reason == c.Expected.Reason && equalIDs(gooo.LocalizedIDs, baseline.LocalizedIDs) && equalIDs(gooo.LocalizedIDs, c.Expected.LocalizedIDs)}
		item.GoooFalsePass, item.GoooFalseNegative = falsePass(gooo, c.Expected), falseNegative(gooo, c.Expected)
		item.BaselineFalsePass, item.BaselineFalseNegative = falsePass(baseline, c.Expected), falseNegative(baseline, c.Expected)
		manifest.Cases = append(manifest.Cases, item)
		manifest.Summary.CaseCount++
		manifest.Summary.ExactOutcomeVector = manifest.Summary.ExactOutcomeVector && item.ExactOutcomeVector
		manifest.Summary.ExactReasonLocalizationVector = manifest.Summary.ExactReasonLocalizationVector && item.ExactReasonVector
		addSummary(&manifest.Summary, gooo, baseline, item)
	}
	if manifest.Summary.BaselineWorkUnits <= manifest.Summary.GoooWorkUnits && manifest.Summary.ExactOutcomeVector && manifest.Summary.ExactReasonLocalizationVector {
		manifest.Finding = NoUniqueBenefit
	} else {
		manifest.Finding = UniqueBenefit
	}
	return manifest, nil
}
func addSummary(s *Summary, gooo, baseline Outcome, item ComparisonCase) {
	s.GoooWorkUnits += gooo.Work.Units
	s.BaselineWorkUnits += baseline.Work.Units
	s.GoooRatio.Selected += gooo.Work.Selected
	s.GoooRatio.Full += gooo.Work.Full
	s.BaselineRatio.Selected += baseline.Work.Selected
	s.BaselineRatio.Full += baseline.Work.Full
	s.GoooProvRecords += gooo.Work.ProvRecords
	s.BaselineProvRecords += baseline.Work.ProvRecords
	s.GoooProvPaths += gooo.Work.ProvPaths
	s.BaselineProvPaths += baseline.Work.ProvPaths
	if item.GoooFalsePass {
		s.GoooFalsePasses++
	}
	if item.GoooFalseNegative {
		s.GoooFalseNegatives++
	}
	if item.BaselineFalsePass {
		s.BaselineFalsePasses++
	}
	if item.BaselineFalseNegative {
		s.BaselineFalseNegatives++
	}
	s.GoooRatio.Known, s.BaselineRatio.Known = s.GoooRatio.Full > 0, s.BaselineRatio.Full > 0
}
func falsePass(got Outcome, want Expected) bool {
	return want.State != ClosedSound && got.State == ClosedSound
}
func falseNegative(got Outcome, want Expected) bool {
	return want.State == ClosedSound && got.State != ClosedSound
}
func equalIDs(left, right []string) bool {
	return len(left) == len(right) && sortedJoin(left) == sortedJoin(right)
}
func sortedJoin(values []string) string {
	result := ""
	for _, value := range sorted(values) {
		result += value + "\x00"
	}
	return result
}
