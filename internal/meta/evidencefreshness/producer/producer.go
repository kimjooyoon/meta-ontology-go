package producer

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

// BuildReceipt is intentionally a bounded observer of the checked-in Gooo
// source. It records identities and boundaries; it does not execute a build or
// grant mutation authority.
func BuildReceipt(source []byte, head string, base model.Context) (model.Receipt, error) {
	if !model.ValidHead(head) {
		return model.Receipt{}, fmt.Errorf("invalid head SHA")
	}
	if err := inspectGooo(source); err != nil {
		return model.Receipt{}, err
	}
	if base.Tuple.Recipe == "" || base.Tuple.Environment == "" || base.Tuple.Runner == "" || base.Tuple.Verifier == "" ||
		base.CurrentEpoch <= 0 || base.CurrentEpoch > baseBoundary(base) || base.EnvironmentBoundary == "" {
		return model.Receipt{}, fmt.Errorf("incomplete base freshness boundary")
	}
	sourceDigest := model.DigestBytes(source)
	tuple := base.Tuple
	tuple.Subject = "subject:gooo/" + sourceDigest
	tuple.Material = "material:source/" + sourceDigest
	return model.SealReceipt(model.Receipt{
		Schema:        model.ReceiptSchema,
		HeadSHA:       head,
		ClaimID:       "gooo://evidence-freshness/claim/checked-source",
		Producer:      model.ProducerID,
		Consumer:      model.ConsumerID,
		MetaOperation: model.MetaOperationID,
		ProofChoice:   model.DefaultProofChoice,
		Tuple:         tuple,
		Boundary: model.TemporalBoundary{ObservationEpoch: base.CurrentEpoch,
			ValidThroughEpoch: baseBoundary(base), EnvironmentBoundary: base.EnvironmentBoundary},
		RepositoryWrites:  0,
		MutationAuthority: false,
	}), nil
}

func baseBoundary(base model.Context) int {
	return base.CurrentEpoch + 7
}

func inspectGooo(source []byte) error {
	packageSeen, namespaceSeen, entityCount, activityCount := false, false, 0, 0
	for lineNumber, raw := range strings.Split(string(source), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "package":
			if len(fields) != 2 {
				return fmt.Errorf("invalid package statement at line %d", lineNumber+1)
			}
			packageSeen = true
		case "namespace":
			if len(fields) != 2 {
				return fmt.Errorf("invalid namespace statement at line %d", lineNumber+1)
			}
			namespaceSeen = true
		case "entity":
			if len(fields) != 4 || fields[2] != "id" || !strings.HasPrefix(fields[3], "\"") || !strings.HasSuffix(fields[3], "\"") {
				return fmt.Errorf("invalid entity statement at line %d", lineNumber+1)
			}
			entityCount++
		case "activity":
			if len(fields) < 5 || !contains(fields, "->") {
				return fmt.Errorf("invalid activity statement at line %d", lineNumber+1)
			}
			activityCount++
		default:
			return fmt.Errorf("unsupported Gooo statement at line %d", lineNumber+1)
		}
	}
	if !packageSeen || !namespaceSeen || entityCount < 6 || activityCount < 3 {
		return fmt.Errorf("Gooo source is incomplete for freshness experiment")
	}
	return nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
