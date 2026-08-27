package proofchoicejudge

import "bytes"

func effectsFor(before, after []byte) effects {
	if before == nil || after == nil {
		return effects{}
	}
	changed := !bytes.Equal(before, after)
	value := 0
	if changed {
		value = 1
	}
	return effects{Observed: true, BeforeStatusDigest: digestBytes(before), AfterStatusDigest: digestBytes(after), RepositoryWrites: value, MutationAuthority: changed}
}
