package evidencequorumwire

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

const (
	Schema                  = "gooo/meta-evidence-quorum-channel/v2"
	CurrentEvidence         = "CURRENT_EVIDENCE"
	SyntheticCounterexample = "SYNTHETIC_COUNTEREXAMPLE"
)

// Receipt is deliberately a raw transport object. It contains no asserted
// origin group, role, confidence, or value. The quorum consumer derives those
// meanings from Channel, Predicate, and the structured provenance below.
type Receipt struct {
	Schema                       string          `json:"schema"`
	EvidenceClass                string          `json:"evidence_class"`
	Channel                      string          `json:"channel"`
	HeadSHA                      string          `json:"head_sha"`
	SourcePath                   string          `json:"source_path"`
	SubjectRawDigest             string          `json:"subject_raw_digest"`
	SubjectSemanticDigest        string          `json:"subject_semantic_digest"`
	PolicySemanticDigest         string          `json:"policy_semantic_digest"`
	ExecutableDigest             string          `json:"executable_digest"`
	DependencyPaths              []string        `json:"dependency_paths"`
	DependencyDigest             string          `json:"dependency_digest"`
	ObservationDigest            string          `json:"observation_digest"`
	SourceExecutionReceiptDigest string          `json:"source_execution_receipt_digest,omitempty"`
	SourceExecutionReceipt       json.RawMessage `json:"source_execution_receipt,omitempty"`
	GeneratedArtifactDigest      string          `json:"generated_artifact_digest,omitempty"`
	Producer                     string          `json:"producer"`
	Consumer                     string          `json:"consumer"`
	MetaOperation                string          `json:"meta_operation"`
	ProofChoice                  string          `json:"proof_choice"`
	Predicate                    string          `json:"predicate"`
	RepositoryWrites             int             `json:"repository_writes"`
	MutationAuthority            bool            `json:"mutation_authority"`
	Digest                       string          `json:"digest"`
}

type observationBody struct {
	EvidenceClass           string `json:"evidence_class"`
	Channel                 string `json:"channel"`
	HeadSHA                 string `json:"head_sha"`
	SourcePath              string `json:"source_path"`
	SubjectRawDigest        string `json:"subject_raw_digest"`
	SubjectSemanticDigest   string `json:"subject_semantic_digest"`
	PolicySemanticDigest    string `json:"policy_semantic_digest"`
	SourceExecutionDigest   string `json:"source_execution_receipt_digest,omitempty"`
	GeneratedArtifactDigest string `json:"generated_artifact_digest,omitempty"`
	Predicate               string `json:"predicate"`
}

func DigestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func DigestJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(raw)
}

func DependencyDigest(paths []string) string {
	ordered := append([]string(nil), paths...)
	sort.Strings(ordered)
	return DigestJSON(ordered)
}

func ObservationDigest(receipt Receipt) string {
	return DigestJSON(observationBody{
		EvidenceClass:           receipt.EvidenceClass,
		Channel:                 receipt.Channel,
		HeadSHA:                 receipt.HeadSHA,
		SourcePath:              receipt.SourcePath,
		SubjectRawDigest:        receipt.SubjectRawDigest,
		SubjectSemanticDigest:   receipt.SubjectSemanticDigest,
		PolicySemanticDigest:    receipt.PolicySemanticDigest,
		SourceExecutionDigest:   receipt.SourceExecutionReceiptDigest,
		GeneratedArtifactDigest: receipt.GeneratedArtifactDigest,
		Predicate:               receipt.Predicate,
	})
}

func Seal(receipt Receipt) Receipt {
	receipt.Schema = Schema
	receipt.Digest = ""
	receipt.Digest = DigestJSON(receipt)
	return receipt
}

func Verify(receipt Receipt) bool {
	digest := receipt.Digest
	receipt.Digest = ""
	return digest != "" && digest == DigestJSON(receipt)
}

func Marshal(receipt Receipt) ([]byte, error) {
	return json.Marshal(Seal(receipt))
}
