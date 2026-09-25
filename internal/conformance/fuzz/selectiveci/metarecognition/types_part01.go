package metarecognition

const SchemaVersion = "gooo/selective-ci-metarecognition/v1"

type Subject string

const (
	SubjectBinding   Subject = "SEMANTIC_BINDING"
	SubjectGraph     Subject = "IMPACT_GRAPH"
	SubjectSoundness Subject = "FULL_SOUNDNESS"
	SubjectPath      Subject = "PATH_CLOSURE"
	SubjectResource  Subject = "RESOURCE_ENVELOPE"
)

func (s Subject) Valid() bool {
	switch s {
	case SubjectBinding, SubjectGraph, SubjectSoundness, SubjectPath, SubjectResource:
		return true
	default:
		return false
	}
}

type State string

const (
	ClosedSound              State = "CLOSED/SOUND"
	FailClosedUnsound        State = "FAIL_CLOSED/UNSOUND"
	UnknownFullSuiteRequired State = "UNKNOWN/FULL_SUITE_REQUIRED"
)

type Reason string

const (
	ReasonExactBinding       Reason = "EXACT_REGISTERED_BINDING"
	ReasonRenameBinding      Reason = "STABLE_ID_RENAME_EXACT_BINDING"
	ReasonBlobWithoutID      Reason = "FILE_BLOB_WITHOUT_STABLE_ID_BINDING"
	ReasonSourceMapRegistry  Reason = "SOURCE_MAP_OR_REGISTRY_MISSING"
	ReasonUnknownGraph       Reason = "UNKNOWN_GRAPH_NODE"
	ReasonMissedObligation   Reason = "MISSED_OBLIGATION"
	ReasonGlobalGuard        Reason = "GLOBAL_GUARD_OMITTED"
	ReasonSelectedDrift      Reason = "SELECTED_FULL_STATUS_OR_OUTPUT_DIGEST_DRIFT"
	ReasonOmittedFailure     Reason = "OMITTED_FULL_SUITE_FAILURE"
	ReasonNonAuthoritative   Reason = "NON_AUTHORITATIVE_OMITTED_OBLIGATION"
	ReasonDuplicateReceipt   Reason = "DUPLICATE_PATH_RECEIPT"
	ReasonConflictingReceipt Reason = "CONFLICTING_PATH_RECEIPT"
	ReasonInvalidResource    Reason = "INVALID_OR_OVERFLOW_RESOURCE_RECEIPT"
	ReasonExternalMissing    Reason = "EXTERNAL_AUTHENTICITY_PROVIDER_PHASE_OBSERVER_MISSING"
)

type Finding string

const (
	NoUniqueBenefit Finding = "NO_UNIQUE_BENEFIT"
	UniqueBenefit   Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
)

type Status string

const (
	Pass Status = "PASS"
	Fail Status = "FAIL"
)

type Authority string

const (
	Authoritative Authority = "AUTHORITATIVE"
	Candidate     Authority = "CANDIDATE"
	Derived       Authority = "DERIVED"
)
