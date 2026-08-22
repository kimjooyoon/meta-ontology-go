package improvement

import "strconv"

const testRegistryDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

func testSnapshot(completed int64) Snapshot {
	evidence := make([]Evidence, SnapshotTotal)
	for index := range evidence {
		status := NotSatisfied
		if int64(index) < completed {
			status = Satisfied
		}
		evidence[index] = Evidence{
			ID:		"obligation-" + strconv.Itoa(index+1),
			Status: status,
		}
	}
	return Snapshot{
		ContractSchema: SnapshotSchema,
		RegistryDigest: testRegistryDigest,
		Completed:		completed,
		Total:			SnapshotTotal,
		BasisPoints:		completed * 10_000 / SnapshotTotal,
		Evidence:		evidence,
	}
}
