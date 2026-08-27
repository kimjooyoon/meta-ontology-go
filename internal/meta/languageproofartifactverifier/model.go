package languageproofartifactverifier

const (
	ReportSchema    = "gooo/language-proof-carrying-artifact-verifier/v1"
	ContractSchema  = "gooo/language-proof-carrying-artifact-contract/v1"
	RecipeSchema    = "gooo/language-proof-carrying-recipe/v1"
	ArtifactSchema  = "gooo/language-proof-carrying-artifact/v1"
	ProducerID      = "gooo://producer/language-proof-carrying-artifact"
	ConsumerID      = "gooo://consumer/language-proof-carrying-artifact-verifier"
	CaseTotal       = 6
	EvidenceTotal   = 3
	TransitionTotal = 4
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type IndependenceEvidence struct {
	Schema                    string `json:"schema"`
	ProducerDependencies      int    `json:"producer_dependencies"`
	ProducerImportNumerator   int    `json:"producer_import_numerator"`
	ProducerImportDenominator int    `json:"producer_import_denominator"`
	CoreParserDependencies    int    `json:"core_parser_dependencies"`
}

type Input struct {
	Contract                 Contract
	HeadSHA                  string
	ValidArtifact            []byte
	TamperedArtifact         []byte
	CoherentTamperedArtifact []byte
	MissingArtifact          []byte
	ByteOnlyArtifact         []byte
	WrongRecipe              []byte
	Source                   []byte
	Operation                []byte
	Recipe                   []byte
	Independence             IndependenceEvidence
	WriteSet                 WriteSetObservation
	CoherentOperation        []byte
	Interventions            []InterventionInput
}

type RecipeStep struct {
	ID            string `json:"id"`
	Input         string `json:"input"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type Recipe struct {
	Schema   string       `json:"schema"`
	Version  int          `json:"version"`
	ID       string       `json:"id"`
	Consumer string       `json:"consumer"`
	Steps    []RecipeStep `json:"steps"`
}

type Evidence struct {
	Kind                            string     `json:"kind"`
	ClaimID                         string     `json:"claim_id"`
	ProofChoice                     string     `json:"proof_choice"`
	MetaOperation                   string     `json:"meta_operation"`
	Coordinate                      Coordinate `json:"coordinate"`
	SourceDigest                    string     `json:"source_digest"`
	SemanticDigest                  string     `json:"semantic_digest"`
	ReceiptDigest                   string     `json:"receipt_digest,omitempty"`
	Activity                        string     `json:"activity,omitempty"`
	Predicate                       string     `json:"predicate,omitempty"`
	RepositoryWrites                int        `json:"repository_writes"`
	MutationAuthority               bool       `json:"mutation_authority"`
	ArtifactUseAuthority            string     `json:"artifact_use_authority"`
	IndependentVerificationRequired bool       `json:"independent_verification_required"`
	EvidenceDigest                  string     `json:"evidence_digest"`
}

type Authority struct {
	ArtifactUseAuthority string `json:"artifact_use_authority"`
	MutationAuthority    bool   `json:"mutation_authority"`
	PromotionAuthority   bool   `json:"promotion_authority"`
	SemanticAuthority    bool   `json:"semantic_authority"`
	Basis                string `json:"basis"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Artifact struct {
	Schema         string              `json:"schema"`
	HeadSHA        string              `json:"head_sha"`
	Producer       string              `json:"producer"`
	Consumer       string              `json:"consumer"`
	MetaOperation  string              `json:"meta_operation"`
	Decision       string              `json:"decision"`
	Resolution     string              `json:"resolution"`
	Reason         string              `json:"reason"`
	SourcePath     string              `json:"source_path"`
	SourceDigest   string              `json:"source_digest"`
	SemanticDigest string              `json:"semantic_digest"`
	Evidence       []Evidence          `json:"evidence"`
	Recipe         Recipe              `json:"recipe"`
	RecipeDigest   string              `json:"recipe_digest"`
	PriorLedger    ClaimLedger         `json:"prior_ledger"`
	WriteSet       WriteSetObservation `json:"write_set"`
	Authority      Authority           `json:"authority"`
	Effects        Effects             `json:"effects"`
	NotClaimed     []string            `json:"not_claimed"`
	Digest         string              `json:"digest"`
}

type CaseSpec struct {
	ID                 string `json:"id"`
	InputKind          string `json:"input_kind"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
}

type Contract struct {
	Schema  string     `json:"schema"`
	Version int        `json:"version"`
	Cases   []CaseSpec `json:"cases"`
}

type WriteSetEntry struct {
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type WriteSetChange struct {
	Path         string `json:"path"`
	BeforeDigest string `json:"before_digest,omitempty"`
	AfterDigest  string `json:"after_digest,omitempty"`
	BeforeKind   string `json:"before_kind,omitempty"`
	AfterKind    string `json:"after_kind,omitempty"`
}

type WriteSetObservation struct {
	Schema            string           `json:"schema"`
	Version           int              `json:"version"`
	Before            []WriteSetEntry  `json:"before"`
	After             []WriteSetEntry  `json:"after"`
	Changed           []WriteSetChange `json:"changed"`
	BeforeDigest      string           `json:"before_digest"`
	AfterDigest       string           `json:"after_digest"`
	RepositoryWrites  int              `json:"repository_writes"`
	MutationAuthority bool             `json:"mutation_authority"`
	Digest            string           `json:"digest"`
}

type LedgerEntry struct {
	ClaimID        string     `json:"claim_id"`
	Status         string     `json:"status"`
	Resolution     string     `json:"resolution"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	ProofChoice    string     `json:"proof_choice"`
	MetaOperation  string     `json:"meta_operation"`
	Coordinate     Coordinate `json:"coordinate"`
	Reason         string     `json:"reason"`
	EvidenceDigest []string   `json:"evidence_digests"`
	Provenance     string     `json:"provenance"`
	PreviousDigest string     `json:"previous_digest"`
	Digest         string     `json:"digest"`
}

type ClaimLedger struct {
	Schema  string        `json:"schema"`
	Version int           `json:"version"`
	Entries []LedgerEntry `json:"entries"`
	Digest  string        `json:"digest"`
}

type SubjectInput struct {
	Artifact  []byte
	Source    []byte
	Operation []byte
	Recipe    []byte
}

type InterventionInput struct {
	ID     string
	Kind   string
	Before SubjectInput
	After  SubjectInput
}

type InterventionResult struct {
	ID                           string `json:"id"`
	Kind                         string `json:"kind"`
	Status                       string `json:"status"`
	Reason                       string `json:"reason"`
	RawSourceDigestBefore        string `json:"raw_source_digest_before"`
	RawSourceDigestAfter         string `json:"raw_source_digest_after"`
	SemanticDigestBefore         string `json:"semantic_digest_before"`
	SemanticDigestAfter          string `json:"semantic_digest_after"`
	OperationReceiptDigestBefore string `json:"operation_receipt_digest_before"`
	OperationReceiptDigestAfter  string `json:"operation_receipt_digest_after"`
	EvidenceLinkDigestBefore     string `json:"evidence_link_digest_before"`
	EvidenceLinkDigestAfter      string `json:"evidence_link_digest_after"`
	ClaimTransitionDigestBefore  string `json:"claim_transition_digest_before"`
	ClaimTransitionDigestAfter   string `json:"claim_transition_digest_after"`
	ConsumerDecisionBefore       string `json:"consumer_decision_before"`
	ConsumerDecisionAfter        string `json:"consumer_decision_after"`
	RawDigestChanged             bool   `json:"raw_digest_changed"`
	SemanticDigestChanged        bool   `json:"semantic_digest_changed"`
	OperationReceiptChanged      bool   `json:"operation_receipt_changed"`
	EvidenceLinksChanged         bool   `json:"evidence_links_changed"`
	ClaimTransitionsChanged      bool   `json:"claim_transitions_changed"`
	SemanticDigestPreserved      bool   `json:"semantic_digest_preserved"`
	ConsumerDecisionPreserved    bool   `json:"consumer_decision_preserved"`
}

type ClaimTransition struct {
	ClaimID        string     `json:"claim_id"`
	Capability     string     `json:"capability"`
	From           string     `json:"from"`
	To             string     `json:"to"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	ProofChoice    string     `json:"proof_choice"`
	MetaOperation  string     `json:"meta_operation"`
	Coordinate     Coordinate `json:"coordinate"`
	Reason         string     `json:"reason"`
	EvidenceDigest []string   `json:"evidence_digests"`
}
