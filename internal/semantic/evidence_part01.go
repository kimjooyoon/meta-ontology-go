package semantic

// Stable host identities let evidence from two compiler implementations be
// compared without treating the implementation itself as semantic meaning.
const (
	GoHostedCompilerID   ID = "gooo://host/compiler/go"
	GoooHostedCompilerID ID = "gooo://host/compiler/gooo"
	GoVerifierID         ID = "gooo://host/verifier/go"
	CIVerifierID            = GoVerifierID

	GoHostedCompiler   = GoHostedCompilerID
	GoooHostedCompiler = GoooHostedCompilerID
	GoVerifier         = GoVerifierID
	CIVerifier         = GoVerifierID
)

// EvidenceKind identifies the producer contract for an evidence record.
type EvidenceKind string

const (
	CompilerRunEvidence  EvidenceKind = "compiler-run"
	VerificationEvidence EvidenceKind = "verification"
	ComparisonEvidence   EvidenceKind = "comparison"
)

func (k EvidenceKind) Valid() bool {
	switch k {
	case CompilerRunEvidence, VerificationEvidence, ComparisonEvidence:
		return true
	default:
		return false
	}
}
func (k EvidenceKind) String() string {
	return string(k)
}

// Evidence is an append-only audit record for a semantic fact. ID is stable
// across hosts; Producer records which host emitted the record. Digest binds
// the claim to the source or verification artifact without changing IR
// semantic equivalence.
type Evidence struct {
	ID       ID
	Producer ID
	Kind     EvidenceKind
	Fact     FactKey
	Status   FactStatus
	Digest   string
	Span     Span
}

func NewEvidence(id, producer ID, kind EvidenceKind, fact FactKey, digest string) (Evidence, error) {
	return (Evidence{ID: id, Producer: producer, Kind: kind, Fact: fact, Digest: digest}).Normalized()
}
