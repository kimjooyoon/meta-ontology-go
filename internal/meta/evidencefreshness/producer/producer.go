package producer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/compiler"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

// BuildReceipt observes the checked-in Gooo source through the canonical
// parser/lowerer. It records raw and semantic material separately and binds
// CI's before/after write-set observation without executing a build or
// granting mutation authority.
func BuildReceipt(source []byte, head string, base model.Context, independence model.IndependenceEvidence, writeSet model.WriteSetObservation) (model.Receipt, error) {
	if !model.ValidHead(head) {
		return model.Receipt{}, fmt.Errorf("invalid head SHA")
	}
	compiled, err := compiler.Compile("examples/evidence-freshness/main.gooo", source)
	if err != nil {
		return model.Receipt{}, err
	}
	if base.Tuple.Recipe == "" || base.Tuple.Environment == "" || base.Tuple.Runner == "" || base.Tuple.Verifier == "" ||
		base.CurrentEpoch <= 0 || base.CurrentEpoch > baseBoundary(base) || base.EnvironmentBoundary == "" {
		return model.Receipt{}, fmt.Errorf("incomplete base freshness boundary")
	}
	rawDigest := model.DigestBytes(source)
	tuple := base.Tuple
	tuple.Subject = "subject:gooo/evidence-freshness/claim"
	tuple.Material = model.MaterialDigest{RawDigest: rawDigest, SemanticDigest: compiled.SemanticDigest}
	return model.SealReceipt(model.Receipt{
		Schema: model.ReceiptSchema, HeadSHA: head,
		ClaimID:  "gooo://evidence-freshness/claim/checked-source",
		Producer: model.ProducerID, Consumer: model.ConsumerID,
		MetaOperation: model.MetaOperationID, ProofChoice: model.DefaultProofChoice,
		SourcePath:   "examples/evidence-freshness/main.gooo",
		PolicyDigest: compiled.PolicyDigest, SourceDigest: rawDigest, SemanticDigest: compiled.SemanticDigest,
		PriorClaimState: model.ClaimOpen, Tuple: tuple,
		Boundary: model.TemporalBoundary{ObservationEpoch: base.CurrentEpoch,
			ValidThroughEpoch: baseBoundary(base), EnvironmentBoundary: base.EnvironmentBoundary},
		Independence: independence, WriteSet: writeSet,
	}), nil
}

func baseBoundary(base model.Context) int { return base.CurrentEpoch + 7 }
