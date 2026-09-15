package metriccounterfactual

import (
	"fmt"
	"os"
	"path/filepath"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func ApplyPlan(root string, plan Plan) ([]Receipt, error) {
	if plan.Schema != PlanSchema || !ValidPlan(plan) {
		return nil, fmt.Errorf("invalid plan")
	}
	receipts := make([]Receipt, 0, len(plan.Mutations))
	for _, mutation := range plan.Mutations {
		receipt, err := applyMutation(root, mutation)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func applyMutation(root string, mutation Mutation) (Receipt, error) {
	native, err := artifact.SafeNative(root, mutation.Path)
	if err != nil {
		return Receipt{}, err
	}
	var before []byte
	switch mutation.Kind {
	case "APPEND":
		before, err = os.ReadFile(native)
	case "CREATE":
		if _, statErr := os.Lstat(native); statErr == nil {
			err = fmt.Errorf("create target exists: %s", mutation.Path)
		} else if !os.IsNotExist(statErr) {
			err = statErr
		}
	default:
		err = fmt.Errorf("unsupported mutation %q", mutation.Kind)
	}
	if err != nil {
		return Receipt{}, err
	}
	after := append(append([]byte(nil), before...), []byte(mutation.Content)...)
	if err := os.MkdirAll(filepath.Dir(native), 0o755); err != nil {
		return Receipt{}, err
	}
	if err := os.WriteFile(native, after, 0o644); err != nil {
		return Receipt{}, err
	}
	beforeDigest := "ABSENT"
	if mutation.Kind != "CREATE" {
		beforeDigest = artifact.ContentDigest(before)
	}
	return Receipt{
		Kind: mutation.Kind, Path: mutation.Path,
		BeforeDigest: beforeDigest, AfterDigest: artifact.ContentDigest(after),
		BeforeLines: artifact.CountLines(before), AfterLines: artifact.CountLines(after),
	}, nil
}
