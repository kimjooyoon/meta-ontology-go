package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeEvidence(settings config, model manifest, head string, restored int, state indexState) error {
	loss := len(model.Entries) - restored
	metrics := []indicator{
		{ID: "materialization.content-loss", Value: loss, Limit: 0, Blocking: true, Consumer: "repository-materializer", Operation: "restore-logical-tree", Proof: "coherence"},
		{ID: "materialization.index-unbound", Value: state.Unbound, Limit: 0, Blocking: true, Consumer: "logical-indexer", Operation: "bind-logical-index", Proof: "foundation"},
		{ID: "materialization.index-dirty", Value: state.Dirty, Limit: 0, Blocking: true, Consumer: "logical-indexer", Operation: "replace-physical-head", Proof: "regression"},
		{ID: "materialization.storage-unbound", Value: state.Unexpected, Limit: 0, Blocking: true, Consumer: "repository-materializer", Operation: "close-physical-storage", Proof: "foundation"},
	}
	report := evidence{Schema: "gooo.repository-materialization-evidence.v1", CurrentSHA: head,
		LogicalOriginSHA: model.SourceSHA, LogicalTreeOID: state.TreeOID,
		ReplacementCommit: state.Replacement, Entries: len(model.Entries), Restored: restored, Indicators: metrics}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	name := filepath.Join(settings.work, "materialization-evidence.json")
	if err := os.WriteFile(name, append(data, '\n'), 0o644); err != nil {
		return err
	}
	for _, metric := range metrics {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking materialization indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
