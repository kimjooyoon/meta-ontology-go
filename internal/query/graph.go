package query

import "sort"

// Graph is a small in-memory relation store. Its zero value is ready for use;
// New is provided for callers that prefer an explicit constructor.
type Graph struct {
	deterministic map[FactKey]Fact
	candidates    map[FactKey]Fact
	binding       *projectionBinding
}

type projectionBinding struct {
	semanticDigest   string
	sourceDigest     string
	evidenceDigest   string
	provenanceDigest string
	sourceStatus     string
	evidenceStatus   string
	provenanceStatus string
}

func New() *Graph {
	graph := &Graph{}
	graph.ensure()
	return graph
}

func (graph *Graph) ensure() {
	if graph.deterministic == nil {
		graph.deterministic = make(map[FactKey]Fact)
	}
	if graph.candidates == nil {
		graph.candidates = make(map[FactKey]Fact)
	}
}

// Add inserts a fact, canonicalizing its IDs and relation. A deterministic
// fact removes a candidate with the same triple; a candidate cannot shadow a
// deterministic fact.
func (graph *Graph) Add(fact Fact) error {
	normalized, err := fact.Normalized()
	if err != nil {
		return err
	}
	graph.ensure()
	key := normalized.Key()
	if normalized.Status == FactCandidate {
		if _, exists := graph.deterministic[key]; exists {
			return nil
		}
		graph.candidates[key] = mergeFacts(graph.candidates[key], normalized)
		return nil
	}
	graph.deterministic[key] = mergeFacts(graph.deterministic[key], normalized)
	delete(graph.candidates, key)
	return nil
}

func (graph *Graph) AddDeterministic(fact Fact) error {
	fact.Status = FactDeterministic
	return graph.Add(fact)
}

func (graph *Graph) AddCandidate(fact Fact) error {
	fact.Status = FactCandidate
	return graph.Add(fact)
}

func (graph Graph) DeterministicFacts() []Fact {
	facts := make([]Fact, 0, len(graph.deterministic))
	for _, fact := range graph.deterministic {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

// Facts is the conventional spelling for the deterministic layer.
func (graph Graph) Facts() []Fact { return graph.DeterministicFacts() }

func (graph Graph) CandidateFacts() []Fact {
	facts := make([]Fact, 0, len(graph.candidates))
	for _, fact := range graph.candidates {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

// Candidates is the conventional spelling for the candidate layer.
func (graph Graph) Candidates() []Fact { return graph.CandidateFacts() }

// Metadata returns a detached, current snapshot of the query projection. The
// graph hash follows the current view, while SemanticDigest remains bound to
// the IR snapshot from which the view was derived.
func (graph Graph) Metadata() ProjectionMetadata {
	metadata := ProjectionMetadata{
		SchemaVersion:    QueryProjectionSchemaVersion,
		GraphHash:        graph.StableHash(),
		SourceStatus:     "unavailable",
		EvidenceStatus:   "unknown",
		ProvenanceStatus: "unknown",
		ProjectionStatus: "unbound",
		AuthorityLabels: []AuthorityLabel{
			{View: ".gooo", Authority: "authoritative", Status: "unavailable"},
			{View: "semantic_ir", Authority: "authoritative", Status: "unavailable"},
			{View: "handwritten_go", Authority: "authoritative", Status: "unavailable"},
			{View: "generated_go", Authority: "derived", Status: "unavailable"},
			{View: "provenance", Authority: "authoritative", Status: "unknown"},
			{View: "query_graph", Authority: "derived", Status: "unbound"},
		},
	}
	if graph.binding == nil {
		return metadata
	}
	metadata.SemanticDigest = graph.binding.semanticDigest
	metadata.SourceDigest = graph.binding.sourceDigest
	metadata.EvidenceDigest = graph.binding.evidenceDigest
	metadata.ProvenanceDigest = graph.binding.provenanceDigest
	metadata.SourceStatus = graph.binding.sourceStatus
	metadata.EvidenceStatus = graph.binding.evidenceStatus
	metadata.ProvenanceStatus = graph.binding.provenanceStatus
	metadata.ProjectionStatus = "derived"
	for index := range metadata.AuthorityLabels {
		switch metadata.AuthorityLabels[index].View {
		case "semantic_ir":
			metadata.AuthorityLabels[index].Status = "bound"
		case "provenance":
			metadata.AuthorityLabels[index].Status = metadata.ProvenanceStatus
		case "query_graph":
			metadata.AuthorityLabels[index].Status = "current"
		}
	}
	return metadata
}

// AllFacts returns a detached deterministic ordering of both fact layers.
func (graph Graph) AllFacts() []Fact {
	facts := append(graph.DeterministicFacts(), graph.CandidateFacts()...)
	sortFacts(facts)
	return facts
}

func (graph Graph) HasFact(key FactKey) bool {
	key, err := normalizeKey(key)
	if err != nil {
		return false
	}
	_, exists := graph.deterministic[key]
	return exists
}

func (graph Graph) HasCandidate(key FactKey) bool {
	key, err := normalizeKey(key)
	if err != nil {
		return false
	}
	_, exists := graph.candidates[key]
	return exists
}

func normalizeKey(key FactKey) (FactKey, error) {
	fact, err := NewFact(key.Subject, key.Predicate, key.Object).Normalized()
	if err != nil {
		return FactKey{}, err
	}
	return fact.Key(), nil
}

func mergeFacts(existing, incoming Fact) Fact {
	if existing.Subject == "" {
		return incoming
	}
	if existing.Reason == "" && incoming.Reason != "" {
		existing.Reason = incoming.Reason
	}
	return existing
}

func sortFacts(facts []Fact) {
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i], facts[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		return left.Reason < right.Reason
	})
}
