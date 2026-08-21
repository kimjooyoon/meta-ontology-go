package freshness

type checker struct {
	snapshot Snapshot
	root     string
	items    []Item
	known    map[string]Kind
	sources  map[string]string
}

// Check evaluates a snapshot without changing it. Every result is derived
// from the snapshot and filesystem contents, then sorted deterministically.
func Check(snapshot Snapshot) Report {
	c := checker{
		snapshot: snapshot,
		root:     snapshot.Root,
		known:    make(map[string]Kind),
		sources:  make(map[string]string),
	}
	if c.root == "" {
		c.root = "."
	}
	c.indexRecords()
	c.checkSources()
	c.checkRequired()
	c.checkArtifacts()
	c.checkEvidence()
	sortItems(c.items)
	return Report{Items: c.items}
}
func (c *checker) indexRecords() {
	for _, source := range c.snapshot.Sources {
		c.index(source.ID, KindSource)
	}
	for _, artifact := range c.snapshot.Artifacts {
		c.index(artifact.ID, artifact.Kind)
	}
	for _, evidence := range c.snapshot.Evidence {
		c.index(evidence.ID, KindEvidence)
	}
}
func (c *checker) index(id string, kind Kind) {
	if id == "" {
		return
	}
	if old, exists := c.known[id]; exists {
		c.add(Item{Kind: kind, ID: id, State: StateInvalid, Detail: "duplicate ID also declared as " + string(old)})
		return
	}
	c.known[id] = kind
}
