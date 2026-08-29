package metrictransition

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"

// Build binds repository measurements to a previously verified meta effect.
func Build(options Options) (Result, error) {
	if err := transformationeffect.VerifyFiles(options.EffectPath, options.ReceiptsPath, options.ProvenancePath, options.PatchPath); err != nil {
		return Result{}, err
	}
	inputs, err := loadInputs(options)
	if err != nil {
		return Result{}, err
	}
	state, err := buildState(inputs)
	if err != nil {
		return Result{}, err
	}
	outcome, err := validateEffectOutcome(inputs)
	if err != nil {
		return Result{}, err
	}
	effect, err := buildEffectEvidence(inputs, outcome)
	if err != nil {
		return Result{}, err
	}
	before := StateReference{Schema: StateSchema, Digest: state.Digest, CommitSHA: options.ExpectedSHA, Materialization: "MEASURED"}
	after := StateReference{Schema: StateSchema, Digest: state.Digest, CommitSHA: options.ExpectedSHA, Materialization: "NO_STATE_ADVANCE", BasisDigest: effect.Artifacts[0].Digest}
	decision, reason := "FAIL_CLOSED", "VERIFIED_MIXED_CLOSED_REFUTED"
	if outcome == effectOutcomeFixedPoint {
		decision, reason = "FIXED_POINT_ZERO_DELTA", "VERIFIED_EFFECT_TREE_IDENTITY"
		after.Materialization = "ALIAS_OF_BEFORE"
	}
	ledger := TransitionLedger{
		Schema: LedgerSchema, Status: "BOUND", Decision: decision, Reason: reason,
		Repository: state.Repository, CommitSHA: options.ExpectedSHA, CIRunID: options.CIRunID,
		Before: before, After: after, Delta: MetricDelta{}, Effect: effect, RootPolicy: state.RootPolicy,
		PromotionAuthorized: false,

		Indicators: transitionIndicators(state, effect, options.ExpectedSHA)}
	ledger, err = sealLedger(ledger)
	if err != nil {
		return Result{}, err
	}
	return Result{State: state, Ledger: ledger}, nil
}

func transitionIndicators(state RepositoryState, effect EffectEvidence, sha string) []Indicator {
	terminalID, terminalOperation, terminalDigest := "coherence.fixed-point-alias", "derive-after-from-tree-identity", effect.Artifacts[0].Digest
	if effect.Outcome != effectOutcomeFixedPoint {
		terminalID, terminalOperation, terminalDigest = "coherence.non-promoting-effect-boundary", "preserve-non-promoting-terminal", effect.SetDigest
	}
	return []Indicator{
		{ID: "foundation.metric-state-schema", Family: "foundation", ProofChoice: "FOUNDATION", Satisfied: true, MetaOperation: "canonicalize-repository-state", EvidenceDigest: state.Digest},
		{ID: "foundation.root-topology-exemption", Family: "foundation", ProofChoice: "FOUNDATION", Satisfied: true, MetaOperation: "exempt-project-root-topology", EvidenceDigest: state.Digest},
		{ID: "coherence.verified-effect-set", Family: "coherence", ProofChoice: "COHERENCE", Satisfied: true, MetaOperation: "bind-transformation-effect", EvidenceDigest: effect.SetDigest},
		{ID: "coherence.exact-head", Family: "coherence", ProofChoice: "COHERENCE", Satisfied: true, MetaOperation: "bind-exact-head", EvidenceDigest: digestBytes([]byte(sha))},
		{ID: terminalID, Family: "coherence", ProofChoice: "COHERENCE", Satisfied: true, MetaOperation: terminalOperation, EvidenceDigest: terminalDigest},
		{ID: "regression.zero-metric-delta", Family: "regression", ProofChoice: "REGRESSION", Satisfied: true, MetaOperation: "terminate-at-fixed-point", EvidenceDigest: state.Digest},
	}
}
