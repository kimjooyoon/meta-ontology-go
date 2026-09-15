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
