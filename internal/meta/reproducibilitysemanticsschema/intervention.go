package reproducibilitysemanticsschema

type InterventionCase struct {
	ID                   string         `json:"id"`
	Kind                 string         `json:"kind"`
	Stage                string         `json:"stage"`
	Step                 string         `json:"step"`
	Reason               string         `json:"reason"`
	SourceDigestBefore   string         `json:"source_digest_before"`
	SourceDigestAfter    string         `json:"source_digest_after"`
	SemanticDigestBefore string         `json:"semantic_digest_before"`
	SemanticDigestAfter  string         `json:"semantic_digest_after"`
	MeaningBefore        Coordinate     `json:"meaning_before"`
	MeaningAfter         Coordinate     `json:"meaning_after"`
	JointBefore          Coordinate     `json:"joint_before"`
	JointAfter           Coordinate     `json:"joint_after"`
	TransitionsBefore    []JudgmentCase `json:"transitions_before"`
	TransitionsAfter     []JudgmentCase `json:"transitions_after"`
	EvidenceDigest       string         `json:"evidence_digest"`
}

type InterventionArtifact struct {
	Schema         string             `json:"schema"`
	Version        int                `json:"version"`
	ContractID     string             `json:"contract_id"`
	SourcePath     string             `json:"source_path"`
	Denominator    int                `json:"denominator"`
	Cases          []InterventionCase `json:"cases"`
	Decision       string             `json:"decision"`
	Resolution     string             `json:"resolution"`
	Reason         string             `json:"reason"`
	Authority      Authority          `json:"authority"`
	ArtifactDigest string             `json:"artifact_digest"`
}
