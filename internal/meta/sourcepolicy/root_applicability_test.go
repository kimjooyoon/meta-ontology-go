package sourcepolicy

import "testing"

func TestProjectRootTopologyHasCatalogProvenApplicability(t *testing.T) {
	report, err := Evaluate(Default(), []Observation{
		{Subject: ".", Dimension: DimensionDirectEntries, Value: 99},
		{Subject: ".", Dimension: DimensionDirectoryKinds, Value: 2},
		{Subject: ".", Dimension: DimensionRootREADME, Value: 0},
		{Subject: "internal/meta", Dimension: DimensionDirectEntries, Value: 11},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Schema != IndicatorSchema || len(report.Indicators) != 4 {
		t.Fatalf("unexpected report: %#v", report)
	}
	rootExemptions := 0
	for _, indicator := range report.Indicators {
		if indicator.Subject == "." {
			rootExemptions++
			if indicator.MetricID == DimensionRootREADME {
				if indicator.ApplicabilityRule != ApplicabilityRuleProjectRootREADME ||
					indicator.ApplicabilityReason != ApplicabilityReasonRootREADMEExempt ||
					indicator.Operation != OperationExemptRootREADME {
					t.Fatalf("invalid root README exemption: %#v", indicator)
				}
				continue
			}
			if indicator.SubjectKind != SubjectKindProjectRoot ||
				indicator.Applicability != ApplicabilityNotApplicable ||
				indicator.ApplicabilityRule != ApplicabilityRuleProjectRootTopology ||
				indicator.ApplicabilityReason != ApplicabilityReasonRootTopologyExempt ||
				indicator.Operation != OperationExemptRoot || indicator.Blocking || !indicator.Satisfied {
				t.Fatalf("invalid root exemption: %#v", indicator)
			}
			continue
		}
		if indicator.SubjectKind != SubjectKindDirectory ||
			indicator.Applicability != ApplicabilityApplicable ||
			indicator.ApplicabilityRule != ApplicabilityRuleDefault ||
			indicator.ApplicabilityReason != ApplicabilityReasonCatalogApplicable ||
			!indicator.Blocking || indicator.Satisfied || indicator.Operation != OperationPartition {
			t.Fatalf("nested directory lost applicable policy: %#v", indicator)
		}
	}
	if rootExemptions != 3 || len(report.Actionable()) != 1 || len(report.Failed()) != 1 {
		t.Fatalf("unexpected applicability routing: %#v", report)
	}
}
