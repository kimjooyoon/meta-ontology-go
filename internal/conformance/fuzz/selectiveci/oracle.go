// Package selectiveci contains a deliberately small, standalone reference
// model for deterministic selective-CI decisions. It is a conformance oracle,
// not a production selector, and must not depend on internal/detection.
package selectiveci

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"strings"
)

// corpusFile is embedded so the conformance corpus is immutable at runtime.
//
//go:embed corpus.json
var corpusFile embed.FS

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

// LoadCorpus returns the embedded canonical corpus.
func LoadCorpus() (Corpus, error) {
	b, err := corpusFile.ReadFile("corpus.json")
	if err != nil {
		return Corpus{}, err
	}
	var corpus Corpus
	if err := json.Unmarshal(b, &corpus); err != nil {
		return Corpus{}, err
	}
	return corpus, nil
}

// CanonicalDigest hashes the order-normalized fixture, excluding its own
// digest field. JSON encoding supplies deterministic object-key ordering.
func CanonicalDigest(c Case) string {
	b, err := json.Marshal(canonicalize(c))
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Evaluate applies the reference rules. Any failure returns a full-suite
// fallback with no partial selection or resource totals.
func Evaluate(c Case) Result {
	digest := CanonicalDigest(c)
	fallback := func(reason Reason) Result {
		return Result{
			Decision:        FullSuiteFallback,
			Reason:          reason,
			CommandIDs:      []string{},
			Argv:            map[string][]string{},
			CanonicalDigest: digest,
		}
	}

	if !c.Evidence.Complete {
		return fallback(incompleteReason)
	}
	if c.Evidence.GlobalGuards {
		return fallback(globalGuardReason)
	}
	if c.Graph.SnapshotID == "" || c.Evidence.SnapshotID == "" || c.Graph.SnapshotID != c.Evidence.SnapshotID {
		return fallback(snapshotReason)
	}

	commands := make(map[string]Command, len(c.Graph.Commands))
	for _, command := range c.Graph.Commands {
		if command.ID == "" {
			return fallback(invalidGraphReason)
		}
		if _, exists := commands[command.ID]; exists {
			return fallback(duplicateIDReason)
		}
		if len(command.Argv) == 0 {
			return fallback(emptyArgvReason)
		}
		for _, arg := range command.Argv {
			if strings.IndexByte(arg, 0) >= 0 {
				return fallback(nulArgvReason)
			}
		}
		commands[command.ID] = command
	}
	if len(commands) == 0 || len(c.Graph.Roots) == 0 {
		return fallback(invalidGraphReason)
	}
	for _, root := range c.Graph.Roots {
		if _, exists := commands[root]; !exists {
			return fallback(danglingCmdReason)
		}
	}

	adjacency := make(map[string][]string, len(commands))
	for _, edge := range c.Graph.Edges {
		if _, exists := commands[edge.From]; !exists {
			return fallback(danglingEdgeReason)
		}
		if _, exists := commands[edge.To]; !exists {
			return fallback(danglingEdgeReason)
		}
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}

	paths := make(map[string]PathEvidence, len(c.Evidence.Paths))
	for _, path := range c.Evidence.Paths {
		if path.Path == "" {
			return fallback(missingPathReason)
		}
		if _, exists := paths[path.Path]; exists {
			return fallback(ambiguousReason)
		}
		paths[path.Path] = path
	}
	if len(c.Evidence.Changes) == 0 {
		return fallback(noChangesReason)
	}

	selected := make(map[string]struct{})
	changedPaths := make(map[string]struct{}, len(c.Evidence.Changes))
	for _, change := range c.Evidence.Changes {
		path, exists := paths[change.Path]
		if !exists {
			return fallback(missingPathReason)
		}
		if path.Stale {
			return fallback(staleReason)
		}
		if !path.Connected {
			return fallback(disconnectedReason)
		}
		if path.Ambiguous || len(path.Owners) > 1 && change.Kind != ChangeDelete {
			return fallback(ambiguousReason)
		}
		if path.Authority != AuthorityAuthoritative {
			return fallback(nonAuthorityReason)
		}
		if len(path.Owners) == 0 {
			return fallback(unknownPathReason)
		}
		if !changeMatchesEvidence(change, path) {
			return fallback(blobMismatchReason)
		}
		changedPaths[change.Path] = struct{}{}
		for _, owner := range path.Owners {
			if _, exists := commands[owner]; !exists {
				return fallback(danglingCmdReason)
			}
			selected[owner] = struct{}{}
		}
	}

	queue := make([]string, 0, len(selected))
	for id := range selected {
		queue = append(queue, id)
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, downstream := range adjacency[id] {
			if _, exists := selected[downstream]; exists {
				continue
			}
			selected[downstream] = struct{}{}
			queue = append(queue, downstream)
		}
	}

	ids := make([]string, 0, len(selected))
	argv := make(map[string][]string, len(selected))
	var cpu, memory uint64
	for id := range selected {
		command := commands[id]
		if math.MaxUint64-cpu < command.CPUUnits {
			return fallback(cpuOverflowReason)
		}
		cpu += command.CPUUnits
		if math.MaxUint64-memory < command.MemoryCeiling {
			return fallback(memoryOverflowReason)
		}
		memory += command.MemoryCeiling
		ids = append(ids, id)
		argv[id] = append([]string(nil), command.Argv...)
	}
	sort.Strings(ids)

	return Result{
		Decision:        Selective,
		Reason:          completeReason,
		CommandIDs:      ids,
		Argv:            argv,
		CPUUnits:        cpu,
		MemoryCeiling:   memory,
		PathCount:       len(changedPaths),
		CanonicalDigest: digest,
	}
}

func changeMatchesEvidence(change PathChange, path PathEvidence) bool {
	switch change.Kind {
	case ChangeDelete:
		return !path.Present && change.Blob == ""
	case ChangeAdd, ChangeModify, ChangeRelocate:
		return path.Present && change.Blob != "" && change.Blob == path.Blob
	default:
		return false
	}
}

func canonicalize(c Case) canonicalCase {
	graph := c.Graph
	graph.Commands = append([]Command(nil), graph.Commands...)
	graph.Edges = append([]Edge(nil), graph.Edges...)
	graph.Roots = append([]string(nil), graph.Roots...)
	sort.Slice(graph.Commands, func(i, j int) bool { return graph.Commands[i].ID < graph.Commands[j].ID })
	sort.Slice(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].From != graph.Edges[j].From {
			return graph.Edges[i].From < graph.Edges[j].From
		}
		return graph.Edges[i].To < graph.Edges[j].To
	})
	sort.Strings(graph.Roots)

	evidence := c.Evidence
	evidence.Paths = append([]PathEvidence(nil), evidence.Paths...)
	evidence.Changes = append([]PathChange(nil), evidence.Changes...)
	for i := range evidence.Paths {
		evidence.Paths[i].Owners = append([]string(nil), evidence.Paths[i].Owners...)
		sort.Strings(evidence.Paths[i].Owners)
	}
	sort.Slice(evidence.Paths, func(i, j int) bool { return evidence.Paths[i].Path < evidence.Paths[j].Path })
	sort.Slice(evidence.Changes, func(i, j int) bool {
		if evidence.Changes[i].Path != evidence.Changes[j].Path {
			return evidence.Changes[i].Path < evidence.Changes[j].Path
		}
		if evidence.Changes[i].Kind != evidence.Changes[j].Kind {
			return evidence.Changes[i].Kind < evidence.Changes[j].Kind
		}
		return evidence.Changes[i].Blob < evidence.Changes[j].Blob
	})

	ids := append([]string(nil), c.Expected.CommandIDs...)
	sort.Strings(ids)
	argv := make(map[string][]string, len(c.Expected.Argv))
	for id, args := range c.Expected.Argv {
		argv[id] = append([]string(nil), args...)
	}

	return canonicalCase{
		Name:      c.Name,
		Partition: c.Partition,
		Graph:     graph,
		Evidence:  evidence,
		Expected: canonicalExpected{
			Decision:      c.Expected.Decision,
			Reason:        c.Expected.Reason,
			CommandIDs:    ids,
			Argv:          argv,
			CPUUnits:      c.Expected.CPUUnits,
			MemoryCeiling: c.Expected.MemoryCeiling,
			PathCount:     c.Expected.PathCount,
		},
	}
}
