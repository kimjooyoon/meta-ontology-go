package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// These formats all consume the original v6 source and v1 declared-case cells.
// A model with another case ABI needs its own source export implementation.
func isContractCandidateSchema(schema string) bool {
	return schema == contractdecision.Schema || schema == contractdecision.PoolingSchema ||
		schema == contractdecision.ChoiceSchema
}

type declaredContractModel interface {
	Fingerprint() string
	ArtifactSchema() string
}

func (m *conditionPathModel) newContractSession(ctx context.Context, prepared *pathplan.PreparedPlan,
	cases []pathplan.TestCase) (*pathplan.ContractSession, error) {
	switch model := m.contract.(type) {
	case *contractdecision.ChoiceModel:
		return prepared.NewChoiceContractSession(ctx, model, cases)
	case *contractdecision.Model:
		return prepared.NewContractSession(ctx, model, cases)
	default:
		return nil, fmt.Errorf("unsupported declared-contract runtime %T", m.contract)
	}
}
