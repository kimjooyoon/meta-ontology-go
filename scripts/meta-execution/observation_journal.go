package main

import (
	"io"
	"os"
)

// Preserve existing source-bound events beside each manifest as they are emitted,
// rather than depending solely on a terminal bundle or a surviving console log.
// A journal is diagnostic only. It does not authorize execution or prove success.
func openObservationJournal(manifestPath string) (*os.File, *metaExecutionTraceState, error) {
	name := manifestPath + ".trace.ndjson"
	if _, err := archivePreviousObservation(name); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, nil, err
	}
	state := newMetaExecutionTraceStateWithWriter(io.MultiWriter(file, os.Stderr))
	state.verifierPackageSummary = newVerifierPackageSummaryCollector(manifestPath)
	return file, state, nil
}
