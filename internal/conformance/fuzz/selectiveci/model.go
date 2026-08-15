// Package selectiveci contains a deliberately small, standalone reference
// model for deterministic selective-CI decisions. It is a conformance oracle,
// not a production selector, and must not depend on internal/detection.
package selectiveci

type Decision string

const (
	Selective            Decision = "SELECTIVE"
	FullSuiteFallback    Decision = "FULL_SUITE_FALLBACK"
	completeReason       Reason   = "COMPLETE_EVIDENCE"
	incompleteReason     Reason   = "INCOMPLETE_EVIDENCE"
	noChangesReason      Reason   = "NO_AUTHORITY_CHANGE"
	globalGuardReason    Reason   = "GLOBAL_GUARD"
	unknownPathReason    Reason   = "UNKNOWN_PATH"
	missingPathReason    Reason   = "MISSING_EVIDENCE_PATH"
	disconnectedReason   Reason   = "DISCONNECTED_EVIDENCE_PATH"
	ambiguousReason      Reason   = "AMBIGUOUS_EVIDENCE_PATH"
	nonAuthorityReason   Reason   = "NON_AUTHORITY_EVIDENCE"
	staleReason          Reason   = "STALE_EVIDENCE"
	blobMismatchReason   Reason   = "BLOB_MISMATCH"
	snapshotReason       Reason   = "SNAPSHOT_MISMATCH"
	duplicateIDReason    Reason   = "DUPLICATE_COMMAND_ID"
	danglingCmdReason    Reason   = "DANGLING_COMMAND"
	danglingEdgeReason   Reason   = "DANGLING_EDGE"
	emptyArgvReason      Reason   = "EMPTY_ARGV"
	nulArgvReason        Reason   = "NUL_ARGV"
	cpuOverflowReason    Reason   = "CPU_OVERFLOW"
	memoryOverflowReason Reason   = "MEMORY_OVERFLOW"
	invalidGraphReason   Reason   = "INVALID_GRAPH"
)

type Reason string

const (
	AuthorityAuthoritative = "authoritative"
	AuthorityCandidate     = "candidate"
	AuthorityDerived       = "derived"

	ChangeAdd      = "add"
	ChangeModify   = "modify"
	ChangeDelete   = "delete"
	ChangeRelocate = "relocate"
)

// Corpus is the checked-in set of independent oracle fixtures.
type Corpus struct {
	SchemaVersion int    `json:"schemaVersion"`
	Cases         []Case `json:"cases"`
}

// Case is one semantic partition and its expected oracle result.
type Case struct {
	Name      string   `json:"name"`
	Partition string   `json:"partition"`
	Graph     Graph    `json:"graph"`
	Evidence  Evidence `json:"evidence"`
	Expected  Expected `json:"expected"`
}

type Graph struct {
	SnapshotID string    `json:"snapshotId"`
	Commands   []Command `json:"commands"`
	Edges      []Edge    `json:"edges"`
	Roots      []string  `json:"roots"`
}

type Command struct {
	ID            string   `json:"id"`
	Argv          []string `json:"argv"`
	CPUUnits      uint64   `json:"cpuUnits"`
	MemoryCeiling uint64   `json:"memoryCeiling"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

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
