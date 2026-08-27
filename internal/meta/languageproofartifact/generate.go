package languageproofartifact

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Generate(input Input) (Artifact, error) {
	if !headPattern.MatchString(input.HeadSHA) || input.SourcePath == "" || len(input.Source) == 0 {
		return Artifact{}, fmt.Errorf("proof-carrying artifact request is incomplete")
	}
	derivedRecipe, err := RecipeFromSource(input.Source)
	if err != nil {
		return Artifact{}, err
	}
	if !sameRecipe(input.Recipe, derivedRecipe) {
		return Artifact{}, fmt.Errorf("proof-carrying recipe drift")
	}
	var receipt sourceexecution.Receipt
	if err := json.Unmarshal(input.Operation, &receipt); err != nil {
		return Artifact{}, fmt.Errorf("decode operation receipt: %w", err)
	}
	if err := sourceexecution.Validate(receipt); err != nil {
		return Artifact{}, fmt.Errorf("validate operation receipt: %w", err)
	}
	sourceDigest := digestBytes(input.Source)
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" ||
		receipt.SourceDigest != sourceDigest || receipt.SemanticDigest == "" || receipt.Filename != input.SourcePath {
		return Artifact{}, fmt.Errorf("operation receipt is not bound to source")
	}
	writeSet, err := normalizeWriteSet(input.WriteSet)
	if err != nil {
		return Artifact{}, err
	}

	evidence := []Evidence{
		{Kind: "SOURCE", ClaimID: "source-bytes-bound", ProofChoice: "FOUNDATION",
			MetaOperation: "bind-source-bytes", Coordinate: Coordinate{"GENERATE", "source-evidence", "SOURCE_BYTES_HASHED"}, SourceDigest: sourceDigest, SemanticDigest: receipt.SemanticDigest},
		{Kind: "OPERATION", ClaimID: "operation-receipt-bound", ProofChoice: "COHERENCE",
			MetaOperation: "bind-operation-receipt", Coordinate: Coordinate{"GENERATE", "operation-evidence", "OPERATION_RECEIPT_LINKED"}, SourceDigest: sourceDigest,
			SemanticDigest: receipt.SemanticDigest, ReceiptDigest: receipt.Digest, Activity: receipt.Entry.Activity},
		{Kind: "INVARIANT", ClaimID: "no-byte-authority", ProofChoice: "REGRESSION",
			MetaOperation: "preserve-no-byte-authority", Coordinate: Coordinate{"GENERATE", "invariant-evidence", "BYTES_DO_NOT_GRANT_AUTHORITY"}, SourceDigest: sourceDigest,
			SemanticDigest: receipt.SemanticDigest, Predicate: "generated-bytes-do-not-grant-authority", RepositoryWrites: writeSet.RepositoryWrites, MutationAuthority: writeSet.MutationAuthority,
			ArtifactUseAuthority: "NONE", IndependentVerificationRequired: true},
	}
	for index := range evidence {
		evidence[index].EvidenceDigest = evidenceDigest(evidence[index])
	}
	recipeDigest := digestValue(input.Recipe)
	claims := []ClaimStatement{
		{ID: "source-bytes-bound", Proposition: "source-bytes-match", TargetDigest: sourceDigest, ProofChoice: "FOUNDATION", MetaOperation: "bind-source-bytes", Coordinate: Coordinate{"GENERATE", "source-evidence", "SOURCE_BYTES_HASHED"}, EvidenceDigest: []string{evidence[0].EvidenceDigest}, State: "CARRIED"},
		{ID: "operation-receipt-bound", Proposition: "operation-receipt-match", TargetDigest: receipt.Digest, Dependencies: []string{"source-bytes-bound"}, ProofChoice: "COHERENCE", MetaOperation: "bind-operation-receipt", Coordinate: Coordinate{"GENERATE", "operation-evidence", "OPERATION_RECEIPT_LINKED"}, EvidenceDigest: []string{evidence[1].EvidenceDigest}, State: "CARRIED"},
		{ID: "no-byte-authority", Proposition: "generated-bytes-do-not-grant-authority", TargetDigest: "READ_ONLY_CONSUMPTION", ProofChoice: "REGRESSION", MetaOperation: "preserve-no-byte-authority", Coordinate: Coordinate{"GENERATE", "invariant-evidence", "BYTES_DO_NOT_GRANT_AUTHORITY"}, EvidenceDigest: []string{evidence[2].EvidenceDigest}, State: "CARRIED"},
		{ID: "recipe-match", Proposition: "consumer-recipe-matches-source-recipe", TargetDigest: recipeDigest, Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority"}, ProofChoice: "COHERENCE", MetaOperation: "match-independent-recipe", Coordinate: Coordinate{"GENERATE", "recipe-evidence", "RECIPE_DIGEST_BOUND"}, EvidenceDigest: []string{evidence[0].EvidenceDigest, evidence[1].EvidenceDigest, evidence[2].EvidenceDigest}, State: "CARRIED"},
		{ID: "consumer-authority", Proposition: "verified-consumer-may-read-only-consume", TargetDigest: "READ_ONLY_CONSUMPTION", Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority", "recipe-match"}, ProofChoice: "COHERENCE", MetaOperation: "grant-read-only-consumption", Coordinate: Coordinate{"GENERATE", "authority-evidence", "AUTHORITY_REQUIRES_INDEPENDENT_RECHECK"}, EvidenceDigest: []string{evidence[0].EvidenceDigest, evidence[1].EvidenceDigest, evidence[2].EvidenceDigest}, State: "NONE"},
	}
	for index := range claims {
		claims[index].Digest = claimStatementDigest(claims[index])
	}
	artifact := Artifact{Schema: ArtifactSchema, HeadSHA: input.HeadSHA, Producer: ProducerID,
		Consumer: ConsumerID, MetaOperation: "emit-proof-carrying-artifact", Decision: ArtifactDecision,
		Resolution: ArtifactResolution, Reason: "PROOF_CARRYING_ARTIFACT_EMITTED", SourcePath: input.SourcePath,
		SourceDigest: sourceDigest, SemanticDigest: receipt.SemanticDigest, OperationDigest: receipt.Digest, Evidence: evidence, Claims: claims, Recipe: input.Recipe,
		RecipeDigest: recipeDigest, PriorLedger: openLedger(claims), WriteSet: writeSet,
		Authority: Authority{ArtifactUseAuthority: "NONE", Basis: "INDEPENDENT_CONSUMER_VERIFICATION_REQUIRED"}, Effects: Effects{RepositoryWrites: writeSet.RepositoryWrites, MutationAuthority: writeSet.MutationAuthority},
		NotClaimed: []string{"consumer authorization", "full compiler semantic correctness", "external side effects"}}
	artifact.Digest = artifactDigest(artifact)
	return artifact, nil
}

func sameRecipe(left, right Recipe) bool {
	return digestValue(left) == digestValue(right)
}
