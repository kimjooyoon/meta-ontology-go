package predecessorbinding

import "fmt"

func Validate(report Report, expectedHead string) error {
	if report.Schema != Schema || report.RegistrySchema != RegistrySchema ||
		report.RegistryDigest != registryDigest() || report.HeadSHA != expectedHead ||
		report.UseCase != UseCase || len(report.Evidence) != Total ||
		report.Summary.Total != Total || len(report.Indicators) != 5 ||
		len(report.Proofs) != 4 || report.RepositoryWrites < 0 {
		return fmt.Errorf("predecessor binding report contract mismatch")
	}
	observations := make([]Observation, 0, len(report.Evidence))
	for _, evidence := range report.Evidence {
		observations = append(observations, evidence.Observation)
	}
	replay := Evaluate(report.HeadSHA, observations, report.RepositoryWrites)
	if replay.ReportDigest != report.ReportDigest {
		return fmt.Errorf("predecessor binding report replay mismatch")
	}
	return nil
}
