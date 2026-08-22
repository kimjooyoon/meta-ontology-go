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

func FoundationBaseline(selected BaselineReference) BaselineReference {
	result := BaselineReference{
		Schema:         BaselineReferenceSchema,
		Workflow:       "Transformation effect ledger",
		RunID:          selected.RunID,
		ArtifactName:   selected.ArtifactName,
		HeadSHA:        selected.HeadSHA,
		FileSHA256:     selected.FileSHA256,
		ArtifactDigest: selected.ArtifactDigest,
		SnapshotDigest: selected.SnapshotDigest,
		RegistryDigest: selected.RegistryDigest,
		Completed:      selected.Completed,
		Total:          selected.Total,
		BasisPoints:    selected.BasisPoints,
	}
	result.Digest = digestJSON(result)
	return result
}
