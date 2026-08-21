package semantic

const (
	AuthoritySource       AuthorityLayer = "SOURCE"
	AuthoritySemantic     AuthorityLayer = "SEMANTIC"
	AuthorityDerived      AuthorityLayer = "DERIVED"
	AuthorityCandidate    AuthorityLayer = "CANDIDATE"
	AuthorityVerification AuthorityLayer = "VERIFICATION"
)

func (l AuthorityLayer) Valid() bool {
	switch l {
	case AuthoritySource, AuthoritySemantic, AuthorityDerived, AuthorityCandidate, AuthorityVerification:
		return true
	default:
		return false
	}
}
func (l AuthorityLayer) String() string { return string(l) }

type AuthorityEffect string

const (
	AuthorityDeclare  AuthorityEffect = "DECLARE"
	AuthorityDerive   AuthorityEffect = "DERIVE"
	AuthorityProject  AuthorityEffect = "PROJECT"
	AuthorityObserve  AuthorityEffect = "OBSERVE"
	AuthorityLift     AuthorityEffect = "LIFT"
	AuthorityVerify   AuthorityEffect = "VERIFY"
	AuthorityDelta    AuthorityEffect = "SEMANTIC_DELTA"
	AuthorityNoChange AuthorityEffect = "NO_SEMANTIC_CHANGE"
)

func (e AuthorityEffect) Valid() bool {
	switch e {
	case AuthorityDeclare, AuthorityDerive, AuthorityProject, AuthorityObserve,
		AuthorityLift, AuthorityVerify, AuthorityDelta, AuthorityNoChange:
		return true
	default:
		return false
	}
}
func (e AuthorityEffect) String() string { return string(e) }

type PhasePlacement struct {
	Phase   InferencePhase
	Ordinal uint64
}
type AuthorityBinding struct {
	Layer  AuthorityLayer
	Effect AuthorityEffect
}
type RuleBinding struct {
	ID      ID
	Version string
	Digest  string
}
type SnapshotDigests struct {
	Source   string
	Semantic string
}
type ProfileBinding struct {
	ID      string
	Version string
	Digest  string
}
type InferenceControls struct {
	CatalogDigest string
	PolicyDigest  string
	Profile       ProfileBinding
}
