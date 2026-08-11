package freshness

import (
	"fmt"
	"os"
	"path/filepath"
)

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
	c := checker{snapshot: snapshot, root: snapshot.Root, known: make(map[string]Kind), sources: make(map[string]string)}
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

func (c *checker) checkSources() {
	for _, source := range c.snapshot.Sources {
		item := Item{Kind: KindSource, ID: source.ID, Path: source.Path, State: StateFresh}
		if source.ID == "" {
			item = invalid(item, "source ID is empty")
			c.add(item)
			continue
		}
		digest := source.Digest
		if source.Path != "" {
			actual, err := c.readDigest(source.Path)
			if err != nil {
				item = stateForRead(item, err)
			} else {
				digest = actual
				if source.Digest != "" && source.Digest != actual {
					item = stale(item, "declared digest does not match source")
				}
			}
		}
		if digest == "" {
			item = invalid(item, "source digest is unavailable")
		} else if !ValidDigest(digest) {
			item = invalid(item, "source digest is invalid")
		}
		if item.State == StateFresh {
			c.sources[source.ID] = digest
		}
		c.add(item)
	}
}

func (c *checker) checkRequired() {
	for _, requirement := range c.snapshot.RequiredArtifacts {
		if !c.hasRecord(requirement.ID, requirement.Kind) {
			c.add(Item{Kind: requirement.Kind, ID: requirement.ID, State: StateMissing, Detail: "required artifact is not declared"})
		}
	}
	for _, requirement := range c.snapshot.RequiredEvidence {
		if !c.hasRecord(requirement.ID, KindEvidence) {
			c.add(Item{Kind: KindEvidence, ID: requirement.ID, State: StateMissing, Detail: "required evidence is not declared"})
		}
	}
}

func (c *checker) hasRecord(id string, kind Kind) bool {
	return id != "" && c.known[id] == kind
}

func (c *checker) checkArtifacts() {
	for _, artifact := range c.snapshot.Artifacts {
		item := c.checkMaterial(artifactResultKind(artifact.Kind), artifact.ID, artifact.Path, artifact.InputIDs, artifact.InputDigest, artifact.ContentDigest)
		item = c.checkArtifactKind(item, artifact.Kind)
		item = c.checkProvenance(item, artifact.ID, artifact.Provenance, false)
		for _, evidenceID := range artifact.EvidenceIDs {
			if !c.hasRecord(evidenceID, KindEvidence) {
				c.add(Item{Kind: KindEvidence, ID: evidenceID, State: StateMissing, Detail: "artifact requires evidence that is not declared"})
			}
		}
		c.add(item)
	}
}

func (c *checker) checkEvidence() {
	for _, evidence := range c.snapshot.Evidence {
		item := c.checkMaterial(KindEvidence, evidence.ID, evidence.Path, evidence.InputIDs, evidence.InputDigest, evidence.ContentDigest)
		item = c.checkProvenance(item, evidence.ID, evidence.Provenance, true)
		c.add(item)
	}
}

func (c *checker) checkMaterial(kind Kind, id, path string, inputIDs []string, inputDigest, contentDigest string) Item {
	item := Item{Kind: kind, ID: id, Path: path, State: StateFresh}
	if id == "" {
		return invalid(item, "record ID is empty")
	}
	if !ValidDigest(inputDigest) {
		item = invalid(item, "record input digest is invalid")
	}
	if !ValidDigest(contentDigest) {
		item = invalid(item, "record content digest is invalid")
	}
	if expected, err := DigestInputs(inputIDs, c.sources); err != nil {
		item = stale(item, "current inputs are unavailable: "+err.Error())
	} else if inputDigest != expected {
		item = stale(item, "input digest does not match current inputs")
	}
	if path != "" {
		actual, err := c.readDigest(path)
		if err != nil {
			item = stateForRead(item, err)
		} else if contentDigest != actual {
			item = stale(item, "content digest does not match file")
		}
	}
	return item
}

func (c *checker) checkArtifactKind(item Item, kind Kind) Item {
	if kind != KindProjection && kind != KindCache {
		return invalid(item, "artifact kind must be generated-projection or cache")
	}
	return item
}

func (c *checker) checkProvenance(item Item, id string, provenance Provenance, required bool) Item {
	if required && (provenance.ActivityID == "" || provenance.EntityID == "") {
		return invalid(item, "evidence provenance requires activity_id and entity_id")
	}
	if !required && (provenance.ActivityID != "" || provenance.EntityID != "") && (provenance.ActivityID == "" || provenance.EntityID == "") {
		return invalid(item, "artifact provenance requires both activity_id and entity_id")
	}
	for _, usedID := range provenance.UsedIDs {
		if usedID == "" || c.known[usedID] == "" {
			return invalid(item, fmt.Sprintf("provenance used ID %q is not declared", usedID))
		}
		if usedID == id {
			return invalid(item, "provenance cannot use itself")
		}
	}
	return item
}

func (c *checker) readDigest(path string) (string, error) {
	resolved := path
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(c.root, filepath.FromSlash(path))
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}
	return HashFile(resolved)
}

func (c *checker) add(item Item) {
	c.items = append(c.items, item)
}

func artifactResultKind(kind Kind) Kind {
	if kind == "" {
		return KindProjection
	}
	return kind
}

func invalid(item Item, detail string) Item {
	item.State = StateInvalid
	item.Detail = detail
	return item
}

func stale(item Item, detail string) Item {
	if item.State == StateInvalid || item.State == StateMissing {
		return item
	}
	item.State = StateStale
	item.Detail = detail
	return item
}

func stateForRead(item Item, err error) Item {
	if item.State == StateInvalid {
		return item
	}
	if os.IsNotExist(err) {
		item.State = StateMissing
		item.Detail = "path is missing"
		return item
	}
	if os.IsPermission(err) {
		return invalid(item, "path is not readable")
	}
	return invalid(item, "cannot read path")
}
