package freshness

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
		c.checkRequirement(requirement, false)
	}
	for _, requirement := range c.snapshot.RequiredEvidence {
		c.checkRequirement(requirement, true)
	}
}
func (c *checker) checkRequirement(requirement Requirement, evidence bool) {
	if requirement.ID == "" {
		c.add(Item{Kind: requirement.Kind, State: StateInvalid, Detail: "required record ID is empty"})
		return
	}
	if evidence {
		if requirement.Kind != KindEvidence {
			c.add(Item{Kind: requirement.Kind, ID: requirement.ID, State: StateInvalid, Detail: "required evidence kind must be evidence"})
			return
		}
	} else if requirement.Kind != KindProjection && requirement.Kind != KindCache {
		c.add(Item{Kind: requirement.Kind, ID: requirement.ID, State: StateInvalid, Detail: "required artifact kind must be generated-projection or cache"})
		return
	}
	if !c.hasRecord(requirement.ID, requirement.Kind) {
		c.add(Item{Kind: requirement.Kind, ID: requirement.ID, State: StateMissing, Detail: "required record is not declared"})
	}
}
func (c *checker) hasRecord(id string, kind Kind) bool {
	recordKind, exists := c.known[id]
	return id != "" && exists && recordKind == kind
}
