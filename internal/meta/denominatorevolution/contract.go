package denominatorevolution

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DecodeContract decodes a verification expectation. It deliberately does not
// manufacture denominator members or cases: those are reconstructed from the
// Gooo source by parseSource and compared with this artifact later.
func DecodeContract(raw []byte) (Contract, error) {
	var contract Contract
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, err
	}
	if contract.Schema != ContractSchema || contract.Version != 1 || contract.Producer == "" || contract.Consumer == "" {
		return Contract{}, fmt.Errorf("DENOMINATOR_EVOLUTION_CONTRACT_IDENTITY_MISMATCH")
	}
	if contract.Denominator.Version == "" || len(contract.Denominator.Obligations) != DenominatorSize || len(contract.Cases) != CaseCount {
		return Contract{}, fmt.Errorf("DENOMINATOR_EVOLUTION_CONTRACT_CARDINALITY_MISMATCH")
	}
	if !contract.Policy.NoAggregateEstimates || len(contract.Policy.ForbiddenClaims) == 0 || len(contract.NotClaimed) == 0 {
		return Contract{}, fmt.Errorf("DENOMINATOR_EVOLUTION_CONTRACT_POLICY_MISMATCH")
	}
	return contract, nil
}
