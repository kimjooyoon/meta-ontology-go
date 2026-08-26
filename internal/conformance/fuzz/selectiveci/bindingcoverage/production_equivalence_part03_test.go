package bindingcoverage

import (
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
)

func compareCounts(oracle Result, output production.Output) []string {
	checks := []string{}
	if uint64(oracle.RequiredBindingCount) != output.RequiredBindingCount {
		checks = append(checks, fmt.Sprintf("required_binding_count=%d want=%d", output.RequiredBindingCount, oracle.RequiredBindingCount))
	}
	if uint64(oracle.PartitionCount) != output.PartitionCount {
		checks = append(checks, fmt.Sprintf("partition_count=%d want=%d", output.PartitionCount, oracle.PartitionCount))
	}
	if uint64(oracle.EndpointReferenceCount) != output.EndpointReferenceCount {
		checks = append(checks, fmt.Sprintf("endpoint_reference_count=%d want=%d", output.EndpointReferenceCount, oracle.EndpointReferenceCount))
	}
	if uint64(oracle.WorkUnits) != output.DeterministicWorkUnits {
		checks = append(checks, fmt.Sprintf("work_units=%d want=%d", output.DeterministicWorkUnits, oracle.WorkUnits))
	}
	return checks
}
func comparePermutation(receipts map[string]productionReceipt) error {
	canonical, ok := receipts["complete-two-bindings"]
	if !ok {
		return fmt.Errorf("missing complete-two-bindings receipt")
	}
	permuted, ok := receipts["permuted-complete-two-bindings"]
	if !ok {
		return fmt.Errorf("missing permuted-complete-two-bindings receipt")
	}
	if canonical.ProductionInputBytes != permuted.ProductionInputBytes || canonical.ProductionInputDigest != permuted.ProductionInputDigest || canonical.ProductionOutputDigest != permuted.ProductionOutputDigest {
		return fmt.Errorf("canonical production permutation changed: canonical=%+v permuted=%+v", canonical, permuted)
	}
	return nil
}
