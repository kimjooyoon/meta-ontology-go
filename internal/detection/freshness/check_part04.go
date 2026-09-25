package freshness

import (
	"fmt"
	"os"
	"path/filepath"
)

func (c *checker) checkProvenance(item Item, id string, provenance Provenance, required bool) Item {
	if required && (provenance.ActivityID == "" || provenance.EntityID == "") {
		return invalid(item, "evidence provenance requires activity_id and entity_id")
	}
	if !required && (provenance.ActivityID != "" || provenance.EntityID != "") && (provenance.ActivityID == "" || provenance.EntityID == "") {
		return invalid(item, "artifact provenance requires both activity_id and entity_id")
	}
	for _, usedID := range provenance.UsedIDs {
		if usedID == "" {
			return invalid(item, "provenance used ID is empty")
		}
		if _, exists := c.known[usedID]; !exists {
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
