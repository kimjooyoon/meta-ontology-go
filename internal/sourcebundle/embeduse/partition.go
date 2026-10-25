package embeduse

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

type SourceImage struct {
	SourceDigest, SemanticDigest string
	Declarations                 []sourcebundle.DeclarationOrigin
}

type ConsumerPartition struct {
	Consumer        int
	State, Reason   string
	Before, After   *SourceImage
	LostIDs, NewIDs []string
}

type PartitionPlan struct {
	Partition     sourcebundle.Partition
	Before, After Report
	Consumers     []ConsumerPartition
}

// PreparePartition proposes one source-set partition and observes its affected
// embed declarations in one package. PRESERVED refers only to the source image
// under sourcebundle.Lower; runtime interpretation and native ABI admission are
// separate. CHANGED and UNASSESSED consumers remain in the plan. No files are
// installed, and no repository-wide completeness or write permission is implied.
func PreparePartition(ctx context.Context, directory string, names []string, filename, beforeID, into string) (PartitionPlan, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	before, err := Inspect(ctx, directory, nil)
	if err != nil {
		return PartitionPlan{}, err
	}
	original, err := captureInputs(before, nil)
	if err != nil {
		return PartitionPlan{}, err
	}
	sources, err := selectedSources(original, names)
	if err != nil {
		return PartitionPlan{}, err
	}
	partition, err := sourcebundle.PreparePartition(sources, filename, beforeID, into)
	if err != nil {
		return PartitionPlan{}, err
	}
	if err := unusedDestination(before.Directory, into); err != nil {
		return PartitionPlan{}, err
	}
	after, err := Inspect(ctx, directory, partition.Sources)
	if err != nil {
		return PartitionPlan{}, err
	}
	if err := samePartitionContext(before, after); err != nil {
		return PartitionPlan{}, err
	}
	projected, err := captureInputs(after, partition.Sources)
	if err != nil {
		return PartitionPlan{}, err
	}
	plan := PartitionPlan{Partition: partition, Before: before, After: after}
	for i := range before.Consumers {
		if consumes(before, i, filename) || consumes(after, i, into) {
			plan.Consumers = append(plan.Consumers, compareConsumer(i, before, after, original, projected))
		}
	}
	if _, err := captureInputs(before, nil); err != nil {
		return PartitionPlan{}, err
	}
	if err := unusedDestination(before.Directory, into); err != nil {
		return PartitionPlan{}, err
	}
	if err := ctx.Err(); err != nil {
		return PartitionPlan{}, err
	}
	return plan, nil
}

func consumes(report Report, index int, name string) bool {
	for _, member := range report.Consumers[index].Members {
		if report.Files[member].Filename == name {
			return true
		}
	}
	return false
}

func compareConsumer(index int, before, after Report, original, projected map[string]string) ConsumerPartition {
	result := ConsumerPartition{Consumer: index, State: "UNASSESSED"}
	var beforeErr, afterErr error
	result.Before, beforeErr = consumerImage(before, index, original)
	result.After, afterErr = consumerImage(after, index, projected)
	if beforeErr != nil || afterErr != nil {
		result.Reason = fmt.Sprintf("source scope could not be lowered: before=%v; after=%v", beforeErr, afterErr)
		return result
	}
	result.LostIDs = differentIDs(result.Before, result.After)
	result.NewIDs = differentIDs(result.After, result.Before)
	result.State = "CHANGED"
	if result.Before.SemanticDigest == result.After.SemanticDigest {
		result.State = "PRESERVED"
	}
	return result
}

func consumerImage(report Report, index int, inputs map[string]string) (*SourceImage, error) {
	names := make([]string, 0, len(report.Consumers[index].Members))
	for _, member := range report.Consumers[index].Members {
		names = append(names, report.Files[member].Filename)
	}
	sources, err := selectedSources(inputs, names)
	if err != nil {
		return nil, err
	}
	bundle, err := sourcebundle.Lower(sources)
	if err != nil {
		return nil, err
	}
	return &SourceImage{bundle.SourceDigest, bundle.IR.StableHash(), bundle.Provenance.Declarations}, nil
}

func differentIDs(left, right *SourceImage) []string {
	present := make(map[string]bool, len(right.Declarations))
	for _, item := range right.Declarations {
		present[item.ID] = true
	}
	var result []string
	for _, item := range left.Declarations {
		if !present[item.ID] {
			result = append(result, item.ID)
		}
	}
	slices.Sort(result)
	return result
}
