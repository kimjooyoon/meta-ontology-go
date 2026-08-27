package proofchoicealgebra

import "bytes"

// ObserveEffects derives the read-only effect from CI's before/after status
// snapshots. A missing snapshot is deliberately not treated as observed.
func ObserveEffects(before, after []byte) Effects {
	if before == nil || after == nil {
		return Effects{}
	}
	changed := !bytes.Equal(before, after)
	return Effects{
		Observed:           true,
		BeforeStatusDigest: digestBytes(before),
		AfterStatusDigest:  digestBytes(after),
		RepositoryWrites:   boolInt(changed),
		MutationAuthority:  changed,
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
