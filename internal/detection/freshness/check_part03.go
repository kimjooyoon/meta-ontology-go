package freshness

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
