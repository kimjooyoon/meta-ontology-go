package main

import "fmt"

type ledgerCounts struct {
	FileWitnesses         int `json:"file_witnesses"`
	GoFiles               int `json:"go_files"`
	GoooFiles             int `json:"gooo_files"`
	OtherFiles            int `json:"other_files"`
	LogicalDirectories    int `json:"logical_directories"`
	StorageDirectories    int `json:"storage_directories"`
	FileSourceBindings    int `json:"file_source_bindings"`
	StorageSourceBindings int `json:"storage_source_bindings"`
	DerivedBindings       int `json:"derived_bindings"`
	SubjectWitnesses      int `json:"subject_witnesses"`
	MetaIndicators        int `json:"meta_indicators"`
}

type witnessLedger struct {
	Schema               string            `json:"schema"`
	Repository           string            `json:"repository"`
	CommitSHA            string            `json:"commit_sha"`
	SourceSchema         string            `json:"source_schema"`
	Policy               sourcePolicy      `json:"policy"`
	PolicyDigest         string            `json:"policy_digest"`
	RootTopologyExempt   bool              `json:"root_topology_exempt"`
	Counts               ledgerCounts      `json:"counts"`
	SubjectWitnessDigest string            `json:"subject_witness_digest"`
	MetaIndicatorDigest  string            `json:"meta_indicator_digest"`
	IndicatorDigest      string            `json:"indicator_digest"`
	SemanticDigest       string            `json:"semantic_digest"`
	Status               string            `json:"status"`
	Indicators           []ledgerIndicator `json:"indicators"`
	Witnesses            []subjectWitness  `json:"witnesses"`
}

func validateLedger(ledger witnessLedger) error {
	if ledger.Schema != "gooo/source-subject-witness-ledger/v1" || ledger.Status != "BOUND" {
		return fmt.Errorf("ledger schema or status is not bound")
	}
	if !ledger.RootTopologyExempt || len(ledger.Witnesses) != ledger.Counts.SubjectWitnesses {
		return fmt.Errorf("ledger root policy or witness count is invalid")
	}
	previous := ""
	for _, witness := range ledger.Witnesses {
		key := witnessKey(witness)
		if previous != "" && key <= previous {
			return fmt.Errorf("witness order is not canonical at %q", key)
		}
		previous = key
		if sealWitness(witness).WitnessDigest != witness.WitnessDigest {
			return fmt.Errorf("witness digest mismatch at %q", key)
		}
	}
	counts := countWitnesses(ledger.Witnesses)
	counts.MetaIndicators = ledger.Counts.MetaIndicators
	if counts != ledger.Counts || digestValues(ledger.Witnesses) != ledger.SubjectWitnessDigest {
		return fmt.Errorf("ledger counts or subject digest mismatch")
	}
	if digestValues(ledger.Indicators) != ledger.IndicatorDigest {
		return fmt.Errorf("ledger indicator digest mismatch")
	}
	for _, indicator := range ledger.Indicators {
		if indicator.Verdict != "PASS" {
			return fmt.Errorf("ledger indicator %q did not pass", indicator.ID)
		}
	}
	if ledgerSemanticDigest(ledger) != ledger.SemanticDigest {
		return fmt.Errorf("ledger semantic digest mismatch")
	}
	return nil
}
