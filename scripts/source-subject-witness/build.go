package main

import (
	"sort"
	"strconv"
)

func buildLedger(report sourceReport, expectedSHA string) (witnessLedger, error) {
	if err := validateSource(report, expectedSHA); err != nil {
		return witnessLedger{}, err
	}
	index := indexIndicators(report.Meta.Indicators)
	rootSummary, err := compileRootSummaryWitness(report.Root, index)
	if err != nil {
		return witnessLedger{}, err
	}
	functions, err := compileFunctionWitnesses(report.Meta.Indicators)
	if err != nil {
		return witnessLedger{}, err
	}
	files := append([]fileMetric(nil), report.Files...)
	logical := append([]directoryMetric(nil), report.Directories...)
	storage := append([]directoryMetric(nil), report.StorageDirectories...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	sort.Slice(logical, func(i, j int) bool { return logical[i].Path < logical[j].Path })
	sort.Slice(storage, func(i, j int) bool { return storage[i].Path < storage[j].Path })
	readmeValue := rootREADMEValue(files)
	witnesses := make([]subjectWitness, 0, len(files)+len(functions)+len(logical)+len(storage)+1)
	for _, file := range files {
		binding, err := fileBinding(file, index)
		if err != nil {
			return witnessLedger{}, err
		}
		witnesses = append(witnesses, fileWitness(file, binding))
	}
	witnesses = append(witnesses, rootSummary)
	witnesses = append(witnesses, functions...)
	for _, directory := range logical {
		witnesses = append(witnesses, directoryWitness("LOGICAL_DIRECTORY", directory, logicalDirectoryBinding()))
	}
	for _, directory := range storage {
		binding, err := storageDirectoryBinding(directory, index, readmeValue)
		if err != nil {
			return witnessLedger{}, err
		}
		witnesses = append(witnesses, directoryWitness("STORAGE_DIRECTORY", directory, binding))
	}
	sort.Slice(witnesses, func(i, j int) bool { return witnessKey(witnesses[i]) < witnessKey(witnesses[j]) })
	counts := countWitnesses(witnesses)
	counts.MetaIndicators = len(report.Meta.Indicators)
	ledger := witnessLedger{Schema: "gooo/source-subject-witness-ledger/v3", Repository: report.Repository, CommitSHA: report.CommitSHA, SourceSchema: report.Meta.Schema, Policy: report.Meta.Policy, PolicyDigest: digestJSON(report.Meta.Policy), RootTopologyExempt: report.Meta.Policy.ExemptProjectRootTopology, RootREADMEExempt: report.Meta.Policy.ExemptProjectRootREADME, Counts: counts, SubjectWitnessDigest: digestValues(witnesses), MetaIndicatorDigest: digestValues(report.Meta.Indicators), Status: "BOUND", Witnesses: witnesses}
	ledger.Indicators = buildLedgerIndicators(ledger)
	ledger.IndicatorDigest = digestValues(ledger.Indicators)
	ledger.SemanticDigest = ledgerSemanticDigest(ledger)
	return ledger, validateLedger(ledger)
}

func itoa(value int) string { return strconv.Itoa(value) }

func ledgerSemanticDigest(ledger witnessLedger) string {
	return digestJSON([]any{ledger.Schema, ledger.Repository, ledger.CommitSHA, ledger.SourceSchema, ledger.PolicyDigest, ledger.RootTopologyExempt, ledger.RootREADMEExempt, ledger.Counts, ledger.SubjectWitnessDigest, ledger.MetaIndicatorDigest, ledger.IndicatorDigest})
}
