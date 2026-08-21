package verify

func newReport(value receipt) (Report, error) {
	report := Report{
		Schema: reportSchema, SubjectSHA: value.SubjectSHA,
		ReceiptDigest: value.Digest, ArtifactID: value.Artifact.ID,
		ArtifactDigest: value.Artifact.Digest, FileCount: 3,
		BindingCount: value.BindingCount, OperationCount: value.OperationCount,
		StepCount: value.StepCount, Status: "VERIFIED",
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	digest, err := digestReport(report)
	if err != nil {
		return Report{}, err
	}
	report.Digest = digest
	return report, nil
}

func digestReport(report Report) (string, error) {
	report.Digest = ""
	return digestValue(report)
}
