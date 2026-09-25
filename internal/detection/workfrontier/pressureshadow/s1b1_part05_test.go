package pressureshadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
	"sort"
)

func setK(input *Input, k uint64) {
	input.Selector.MinimumSelectedPressures = uint32(k)
	for index := range input.PathCoverage {
		input.PathCoverage[index].Coverage.RequestedK = k
	}
}
func setK21(input *Input) {
	for number := 4; number <= 21; number++ {
		id := fmt.Sprintf("p-%02d", number)
		input.Selector.Pressures = append(input.Selector.Pressures, workfrontier.Pressure{StableID: id})
		for index := range input.Selector.Paths {
			input.Selector.Paths[index].RequiredPressureIDs = append(input.Selector.Paths[index].RequiredPressureIDs, id)
		}
		for index := range input.PathCoverage {
			coverage := &input.PathCoverage[index].Coverage
			coverage.RequiredPressureIDs = append(coverage.RequiredPressureIDs, id)
			coverage.PressureRecords = append(coverage.PressureRecords, pressurecoverage.PressureRecord{
				PressureID: id, CategoryID: "category-" + id, IndependenceGroupID: "group-" + id, ApplicabilityRuleID: "rule-1",
			})
		}
	}
	setK(input, 21)
	for _, id := range []string{"path/a", "path/b", "path/c"} {
		rebindCoverage(input, id)
	}
}
func rebindCoverage(input *Input, pathID string) {
	row := b2Coverage(input, pathID)
	unsigned := row.Coverage
	unsigned.AuthoritySnapshotDigest, unsigned.PolicyDigest = "", ""
	unsigned.RegistryDigest, unsigned.ToolchainOptionsDigest = "", ""
	unsigned.PressureRecords = append([]pressurecoverage.PressureRecord{}, unsigned.PressureRecords...)
	unsigned.RequiredPressureIDs = append([]string{}, unsigned.RequiredPressureIDs...)
	sort.Slice(unsigned.PressureRecords, func(left, right int) bool {
		return unsigned.PressureRecords[left].PressureID < unsigned.PressureRecords[right].PressureID
	})
	sort.Strings(unsigned.RequiredPressureIDs)
	data, _ := json.Marshal(unsigned)
	inputDigest := testDigest(data)
	row.Coverage.AuthoritySnapshotDigest = testRoleDigest("authority-snapshot", inputDigest)
	row.Coverage.PolicyDigest = testRoleDigest("policy", inputDigest)
	row.Coverage.RegistryDigest = testRoleDigest("registry", inputDigest)
	row.Coverage.ToolchainOptionsDigest = testRoleDigest("toolchain-options", inputDigest)
}
func testRoleDigest(role, inputDigest string) string {
	return testDigest([]byte(role + "\x00" + inputDigest))
}
func testDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
