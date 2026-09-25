package selectiveci

type Evidence struct {
	SnapshotID   string         `json:"snapshotId"`
	Complete     bool           `json:"complete"`
	GlobalGuards bool           `json:"globalGuards"`
	Paths        []PathEvidence `json:"paths"`
	Changes      []PathChange   `json:"changes"`
}
type PathEvidence struct {
	Path      string   `json:"path"`
	Owners    []string `json:"owners"`
	Blob      string   `json:"blob"`
	Present   bool     `json:"present"`
	Connected bool     `json:"connected"`
	Ambiguous bool     `json:"ambiguous"`
	Stale     bool     `json:"stale"`
	Authority string   `json:"authority"`
}
type PathChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Blob string `json:"blob"`
}
type Expected struct {
	Decision        Decision            `json:"decision"`
	Reason          Reason              `json:"reason"`
	CommandIDs      []string            `json:"commandIds"`
	Argv            map[string][]string `json:"argv"`
	CPUUnits        uint64              `json:"cpuUnits"`
	MemoryCeiling   uint64              `json:"memoryCeiling"`
	PathCount       int                 `json:"pathCount"`
	CanonicalDigest string              `json:"canonicalDigest"`
}

// Result is the complete observable output of the oracle.
type Result struct {
	Decision        Decision            `json:"decision"`
	Reason          Reason              `json:"reason"`
	CommandIDs      []string            `json:"commandIds"`
	Argv            map[string][]string `json:"argv"`
	CPUUnits        uint64              `json:"cpuUnits"`
	MemoryCeiling   uint64              `json:"memoryCeiling"`
	PathCount       int                 `json:"pathCount"`
	CanonicalDigest string              `json:"canonicalDigest"`
}
type canonicalExpected struct {
	Decision      Decision            `json:"decision"`
	Reason        Reason              `json:"reason"`
	CommandIDs    []string            `json:"commandIds"`
	Argv          map[string][]string `json:"argv"`
	CPUUnits      uint64              `json:"cpuUnits"`
	MemoryCeiling uint64              `json:"memoryCeiling"`
	PathCount     int                 `json:"pathCount"`
}
type canonicalCase struct {
	Name      string            `json:"name"`
	Partition string            `json:"partition"`
	Graph     Graph             `json:"graph"`
	Evidence  Evidence          `json:"evidence"`
	Expected  canonicalExpected `json:"expected"`
}
