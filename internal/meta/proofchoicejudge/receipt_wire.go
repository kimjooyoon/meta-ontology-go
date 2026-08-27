package proofchoicejudge

type receipt struct {
	Schema               string         `json:"schema"`
	Decision             string         `json:"decision"`
	Reason               string         `json:"reason"`
	SubjectResolution    string         `json:"subject_resolution"`
	SourcePath           string         `json:"source_path"`
	SourceDigest         string         `json:"source_digest"`
	SemanticDigest       string         `json:"semantic_digest"`
	SourceReconstruction reconstruction `json:"source_reconstruction"`
	Items                []item         `json:"items"`
	Transitions          []transition   `json:"transitions"`
	Summary              summary        `json:"summary"`
	Effects              effects        `json:"effects"`
	Digest               string         `json:"digest"`
}

type Verdict struct {
	Schema                 string         `json:"schema"`
	Decision               string         `json:"decision"`
	Reason                 string         `json:"reason"`
	ReceiptDigest          string         `json:"receipt_digest"`
	ComputedReceiptDigest  string         `json:"computed_receipt_digest"`
	ReceiptDigestMatch     bool           `json:"receipt_digest_match"`
	SourceDigest           string         `json:"source_digest"`
	ComputedSourceDigest   string         `json:"computed_source_digest"`
	SourceDigestMatch      bool           `json:"source_digest_match"`
	SemanticDigest         string         `json:"semantic_digest"`
	ComputedSemanticDigest string         `json:"computed_semantic_digest"`
	SemanticDigestMatch    bool           `json:"semantic_digest_match"`
	EffectsMatch           bool           `json:"effects_match"`
	SourceReconstruction   reconstruction `json:"source_reconstruction"`
	Items                  int            `json:"items"`
	Transitions            int            `json:"transitions"`
	IndependentEvidence    int            `json:"independent_evidence"`
	ReceiptOnly            bool           `json:"receipt_only"`
}
