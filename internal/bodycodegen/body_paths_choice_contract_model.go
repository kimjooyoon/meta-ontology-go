package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Each format has a matching source and declared-case export implementation.
func isContractCandidateSchema(schema string) bool {
	return schema == contractdecision.Schema || schema == contractdecision.PoolingSchema ||
		schema == contractdecision.ChoiceSchema || schema == contractdecision.InteractionRequirementSchema ||
		schema == contractdecision.OrderedRequirementSchema
}

func (m *conditionPathModel) orderedContract() bool {
	return m.contract != nil && (m.contract.ArtifactSchema() == contractdecision.InteractionRequirementSchema ||
		m.contract.ArtifactSchema() == contractdecision.OrderedRequirementSchema)
}

type declaredContractModel interface {
	Fingerprint() string
	ArtifactSchema() string
}

func (m *conditionPathModel) newContractSession(ctx context.Context, prepared *pathplan.PreparedPlan,
	cases []pathplan.TestCase) (*pathplan.ContractSession, error) {
	switch model := m.contract.(type) {
	case *contractdecision.InteractionRequirementModel:
		return prepared.NewInteractionRequirementContractSession(ctx, model, cases)
	case *contractdecision.OrderedRequirementModel:
		return prepared.NewOrderedRequirementContractSession(ctx, model, cases)
	case *contractdecision.ChoiceModel:
		return prepared.NewChoiceContractSession(ctx, model, cases)
	case *contractdecision.Model:
		return prepared.NewContractSession(ctx, model, cases)
	default:
		return nil, fmt.Errorf("unsupported declared-contract runtime %T", m.contract)
	}
}
