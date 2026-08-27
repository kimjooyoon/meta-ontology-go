package metainvocation

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema {
		return fmt.Errorf("report schema %q is not supported", report.Schema)
	}
	if report.Decision != DecisionPass && report.Decision != DecisionClosed && report.Decision != DecisionUnknown {
		return fmt.Errorf("report decision %q is not recognized", report.Decision)
	}
	if report.Decision == DecisionUnknown && report.Resolution != ResolutionLower {
		return fmt.Errorf("unknown decision must lower resolution")
	}
	if report.Decision != DecisionUnknown && report.Resolution != ResolutionExact {
		return fmt.Errorf("known decision must retain exact resolution")
	}
	if len(report.Claims) != 3 {
		return fmt.Errorf("report has %d persistent claims, want 3", len(report.Claims))
	}
	if report.Effects.RepositoryWrites != 0 || report.Effects.MutationAuthority {
		return fmt.Errorf("meta invocation exceeded zero-effect boundary")
	}
	if report.Plan.Digest != sealPlan(report.Plan).Digest {
		return fmt.Errorf("plan digest mismatch")
	}
	if report.Receipt.Digest != sealReceipt(report.Receipt).Digest {
		return fmt.Errorf("verification receipt digest mismatch")
	}
	if report.ReportDigest != sealReport(report).ReportDigest {
		return fmt.Errorf("report digest mismatch")
	}
	return nil
}
