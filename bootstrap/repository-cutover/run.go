package main

import "fmt"

func execute(input cutoverConfig) error {
	settings, err := resolveConfig(input)
	if err != nil {
		return err
	}
	model, expected, authoritySHA, candidateSHA, authorityDrift, err := readCatalog(settings)
	if err != nil {
		return err
	}
	physical, err := leafPaths(settings.physical)
	if err != nil {
		return err
	}
	git, err := inspectGit(settings)
	if err != nil {
		return err
	}
	identityDrift := 0
	if git.Head != settings.expectedSHA {
		identityDrift = 1
	}
	report := cutoverEvidence{
		Schema: "gooo.repository-cutover.v1", SourceSHA: settings.expectedSHA,
		LogicalOriginSHA: model.SourceSHA, AuthoritySHA256: authoritySHA,
		CandidateSHA256: candidateSHA, LogicalEntries: len(git.Tracked),
		PhysicalEntries: len(physical), PlannedPaths: unionCount(git.Tracked, expected),
		Indicators: []cutoverIndicator{
			metric("cutover.authority-drift", "bind-ci-manifest", "foundation", authorityDrift, true),
			metric("cutover.storage-unbound", "close-physical-path-set", "coherence", mismatch(expected, physical), true),
			metric("cutover.identity-drift", "bind-source-identity", "foundation", identityDrift, true),
			metric("cutover.workspace-dirty", "require-clean-source", "regression", git.Dirty, true),
			metric("cutover.stage-unbound", "bind-physical-index", "coherence", 0, settings.apply),
		},
	}
	if settings.apply && model.SourceSHA != settings.expectedSHA {
		report.Indicators[2].Value++
	}
	if err := requirePassing(report); err != nil {
		_ = writeEvidence(settings.evidence, report)
		return err
	}
	if settings.apply {
		if err := applyCutover(settings); err != nil {
			return err
		}
		unbound, err := stageCutover(settings, git.Tracked, expected)
		if err != nil {
			return err
		}
		report.Applied = true
		report.Indicators[4].Value = unbound
	}
	if err := writeEvidence(settings.evidence, report); err != nil {
		return err
	}
	if err := requirePassing(report); err != nil {
		return err
	}
	fmt.Printf("repository-cutover: apply=%t logical=%d physical=%d paths=%d\n",
		report.Applied, report.LogicalEntries, report.PhysicalEntries, report.PlannedPaths)
	return nil
}
