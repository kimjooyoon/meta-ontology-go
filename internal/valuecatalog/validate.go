package valuecatalog

import (
	"fmt"
	"regexp"
)

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Validate(report Report, headSHA string) error {
	if report.Schema != ReportSchema || report.HeadSHA != headSHA || !commitPattern.MatchString(headSHA) {
		return fmt.Errorf("catalog identity is invalid")
	}
	if report.Resolution != ResolutionCoreValue || report.SourceLines != 5 || report.ActivitiesObserved != 2 {
		return fmt.Errorf("catalog source boundary is not exact")
	}
	if !validDigest(report.SourceDigest) || !validDigest(report.BeforeSourceDigest) || len(report.CoreIRFingerprint) != 64 {
		return fmt.Errorf("catalog digests are invalid")
	}
	if report.Baseline.Activity != BaselineActivity || report.Baseline.Program != "int.add:1" || report.Baseline.Passed != 3 {
		return fmt.Errorf("baseline program is not exact")
	}
	if report.BaselineCoreProgram != "int.add:1" || report.Improvement.ID != "source-only-catalog-extension" || report.Improvement.Before != coordinate(0, 1) {
		return fmt.Errorf("baseline coordinate is not exact")
	}
	if len(report.Indicators) != CatalogIndicatorCount || len(report.Views) != 3 || len(report.Proofs) != 3 {
		return fmt.Errorf("catalog denominator changed")
	}
	if err := validateOperationSpecEvidence(report); err != nil {
		return err
	}
	if err := validateCatalogState(report); err != nil {
		return err
	}
	if report.Summary.RepositoryWrites != 0 || report.Authority.RepositoryMutationAuthorized || report.Authority.PromotionAuthorized || report.Authority.AutomaticAdoptionAuthorized {
		return fmt.Errorf("catalog authority boundary changed")
	}
	if len(report.NonClaims) != 5 || report.Digest != reportDigest(report) {
		return fmt.Errorf("catalog non-claim or digest changed")
	}
	return nil
}
