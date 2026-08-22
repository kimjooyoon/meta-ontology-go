package proposalpromotion

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposal"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

func Build(
	currentHead, evidenceHead string,
	selection proposalpredecessor.Report, contractRaw []byte,
) (Receipt, error) {
	if err := proposalpredecessor.Validate(selection); err != nil {
		return Receipt{}, fmt.Errorf("validate proposal predecessor: %w", err)
	}
	selectionRaw, err := json.Marshal(selection)
	if err != nil {
		return Receipt{}, err
	}
	selectionData := selectionView{}
	if err := json.Unmarshal(selectionRaw, &selectionData); err != nil {
		return Receipt{}, err
	}
	contract := proposal.Report{}
	if err := json.Unmarshal(contractRaw, &contract); err != nil {
		return Receipt{}, fmt.Errorf("decode proposal contract: %w", err)
	}
	if err := proposal.Validate(contract); err != nil {
		return Receipt{}, fmt.Errorf("validate proposal contract: %w", err)
	}
	contractData := contractView{}
	if err := json.Unmarshal(contractRaw, &contractData); err != nil {
		return Receipt{}, err
	}
	if err := validateLinkage(currentHead, evidenceHead, selectionData, contractData, contractRaw); err != nil {
		return Receipt{}, err
	}
	receipt := evaluate(currentHead, evidenceHead, sourceFrom(selectionData, contractData, contractRaw))
	if err := Validate(receipt, currentHead); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func validateLinkage(
	currentHead, evidenceHead string, selection selectionView,
	contract contractView, contractRaw []byte,
) error {
	if selection.CurrentSubjectSHA != currentHead || selection.PredecessorSHA != evidenceHead ||
		selection.Selected.HeadSHA != evidenceHead || contract.SubjectSHA != evidenceHead {
		return fmt.Errorf("FAIL_CLOSED: proposal promotion subject linkage mismatch")
	}
	if selection.Selected.ProposalFileSHA256 != digestBytes(contractRaw) ||
		selection.Selected.ProposalReportDigest != contract.ReportDigest {
		return fmt.Errorf("FAIL_CLOSED: proposal promotion digest linkage mismatch")
	}
	return nil
}
