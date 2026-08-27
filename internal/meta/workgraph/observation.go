package workgraph

type Observation struct {
	HeadSHA           string
	SourcePath        string
	SourceText        string
	SourceDigest      string
	CheckDigest       string
	GeneratedDigest   string
	ReplayDigest      string
	GeneratedBytes    int64
	Resource          ResourceSample
	Predecessor       *Report
	PredecessorDigest string
}
