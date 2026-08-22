package artifact

const BaselineReferenceSchema = "gooo/language-readiness-baseline-reference/v1"

type BaselineReference struct {
	Schema         string `json:"schema"`
	Workflow       string `json:"workflow"`
	RunID          int64  `json:"run_id"`
	ArtifactName   string `json:"artifact_name"`
	HeadSHA        string `json:"head_sha"`
	FileSHA256     string `json:"file_sha256"`
	ArtifactDigest string `json:"artifact_digest"`
	SnapshotDigest string `json:"snapshot_digest"`
	RegistryDigest string `json:"registry_digest"`
	Completed      int    `json:"completed"`
	Total          int    `json:"total"`
	BasisPoints    int    `json:"basis_points"`
	Digest         string `json:"digest"`
}

func FoundationBaseline() BaselineReference {
	result := BaselineReference{
		Schema:   BaselineReferenceSchema,
		Workflow: "Transformation effect ledger",
		RunID:    32582965541,
		ArtifactName: "language-readiness-artifact-" +
			"fb893cf61db1896c7f9617ae54a8befd2d232f33",
		HeadSHA:        "fb893cf61db1896c7f9617ae54a8befd2d232f33",
		FileSHA256:     "sha256:6c0b35ea341674f60d754c4ac666b89d136c3c90cdfb2d83696ef8acf0d61bb5",
		ArtifactDigest: "sha256:d7f3755aa9f8b30467f71cce498203f5253180a631347b1a6a2f5dd43b0c1f9a",
		SnapshotDigest: "sha256:0cfc4b29160d19408dae4113a7d33f486c6c71c317d4560a71e1858097f64375",
		RegistryDigest: "sha256:08612bd4991d724bb65a093d68763c3af0ac782a9ab58f9b907e818cb57ba05c",
		Completed:      7,
		Total:          24,
		BasisPoints:    2916,
	}
	result.Digest = digestJSON(result)
	return result
}
