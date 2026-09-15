package proposalcompat

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
)

func Build(raw []byte, expectedRepository, expectedHead, expectedPredecessor string) (Bundle, error) {
	var current proposalpromotion.Receipt
	if err := json.Unmarshal(raw, &current); err != nil {
		return Bundle{}, fmt.Errorf("decode v2 proposal promotion: %w", err)
	}
	if err := proposalpromotion.Validate(current, expectedRepository, expectedHead, expectedPredecessor); err != nil {
		return Bundle{}, err
	}
	legacy := sealLegacy(LegacyReceipt{Schema: LegacySchema,
		CurrentHeadSHA: current.CurrentHeadSHA, Decision: current.Decision,
		Summary: LegacySummary{Satisfied: current.Summary.Satisfied,
			Total: current.Summary.Total, Unresolved: current.Summary.Unresolved,
			RepositoryWrites: current.Summary.RepositoryWrites}})
	legacyPayload := EncodeLegacy(legacy)
	source := Source{ExpectedHeadSHA: expectedHead, SourceSchema: current.Schema,
		SourceDecision: current.Decision, SourceReportDigest: current.ReportDigest,
		SourceFileSHA256: digestBytes(raw), SourceSatisfied: current.Summary.Satisfied,
		SourceTotal: current.Summary.Total, SourceUnresolved: current.Summary.Unresolved,
		SourceRepositoryWrites:   current.RepositoryWrites,
		SourceMutationAuthorized: current.RepositoryMutationAuthorized,
		TargetSchema:             LegacySchema, TargetReportDigest: legacy.ReportDigest,
		TargetFileSHA256: digestBytes(legacyPayload), ProjectedFields: projectedFields}
	bundle := Bundle{Legacy: legacy, Receipt: buildReceipt(source)}
	return bundle, Validate(bundle, expectedHead)
}
