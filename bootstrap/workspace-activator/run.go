package main

import "fmt"

func execute(input activationConfig) error {
	settings, err := resolveConfig(input)
	if err != nil {
		return err
	}
	source, err := readSourceEvidence(settings.materialization, settings.expectedSHA)
	if err != nil {
		return err
	}
	physical, err := leafCount(settings.root)
	if err != nil {
		return err
	}
	logical, err := leafCount(settings.logical)
	if err != nil {
		return err
	}
	if physical == 0 || logical != source.Entries {
		return fmt.Errorf("activation input cardinality is not exact")
	}
	if err := requireEmpty(settings.storage); err != nil {
		return err
	}
	if err := moveChildren(settings.root, settings.storage, true); err != nil {
		return err
	}
	if err := moveChildren(settings.logical, settings.root, false); err != nil {
		return err
	}
	stored, err := leafCount(settings.storage)
	if err != nil {
		return err
	}
	activated, err := leafCount(settings.root)
	if err != nil {
		return err
	}
	dirty, drift, err := inspectGit(settings)
	if err != nil {
		return err
	}
	report := activationEvidence{
		Schema: "gooo.workspace-activation.v1", SourceSHA: settings.expectedSHA,
		PhysicalEntries: physical, StoredEntries: stored,
		LogicalEntries: logical, ActivatedEntries: activated,
		Indicators: []activationIndicator{
			metric("activation.storage-loss", "preserve-physical-tree", "foundation", distance(stored, physical)),
			metric("activation.logical-loss", "activate-logical-tree", "coherence", distance(activated, logical)),
			metric("activation.git-dirty", "bind-active-index", "regression", dirty),
			metric("activation.identity-drift", "preserve-commit-identity", "foundation", drift),
		},
	}
	if err := writeActivationEvidence(settings.evidence, report); err != nil {
		return err
	}
	if err := requirePassing(report); err != nil {
		return err
	}
	fmt.Printf("workspace-activator: physical=%d logical=%d dirty=%d drift=%d\n", stored, activated, dirty, drift)
	return nil
}
