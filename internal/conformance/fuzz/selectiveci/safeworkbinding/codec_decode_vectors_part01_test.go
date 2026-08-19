package safeworkbinding

type decodeResultWant struct {
	decision          Decision
	reason            Reason
	fault             Reason
	emptyFaults       bool
	fullSuiteRequired bool
	resultDigest      Digest
	replayDigest      Digest
	resultFrameLength int
}
type decodeVector struct {
	name    string
	input   []byte
	binding SafeWorkBinding
	want    decodeResultWant
}
