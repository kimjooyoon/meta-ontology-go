package selectiveci

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// corpusFile is embedded so the conformance corpus is immutable at runtime.
//
//go:embed corpus.json
var corpusFile embed.FS

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

func canonicalize(c Case) canonicalCase {
	graph := canonicalGraph(c.Graph)
	evidence := canonicalEvidence(c.Evidence)
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

func canonicalGraph(graph Graph) Graph {
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
	return graph
}

func canonicalEvidence(evidence Evidence) Evidence {
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
	return evidence
}
