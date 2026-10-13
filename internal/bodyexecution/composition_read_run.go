package bodyexecution

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type compositionReadRun struct {
	construction          *compositionConstructionContext
	first                 []byte
	rows                  [][]CompositionReadObservation
	dropped, registration []int
}

func (run *compositionReadRun) reset(root string, r *CompositionRuntime) error {
	if len(r.ReadSites)+len(r.ConstructionSites) == 0 {
		return nil
	}
	err := os.Remove(filepath.Join(root, readObservationFile))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (run *compositionReadRun) observe(attempt int, root string, graph compositionGraph, suite CompositionCases, r *CompositionRuntime) error {
	if len(r.ReadSites)+len(r.ConstructionSites) == 0 {
		return nil
	}
	file, err := os.Open(filepath.Join(root, readObservationFile))
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, (32<<20)+1))
	if err != nil {
		return err
	}
	if len(raw) > 32<<20 {
		return fmt.Errorf("native collection reads exceed 32 MiB")
	}
	if attempt != 0 {
		if !bytes.Equal(run.first, raw) {
			return fmt.Errorf("compiled collection read origins differ between runs")
		}
		return nil
	}
	run.first = raw
	if len(r.ConstructionSites) > 0 {
		run.construction = &compositionConstructionContext{sites: r.ConstructionSites}
	}
	run.rows, run.dropped, run.registration, r.ReadObservation, err = graph.nativeReadFile(raw, suite, r.ReadSites, r.ReadInputs, run.construction)
	return err
}

func (run *compositionReadRun) bind(graph compositionGraph, traces []CompositionTrace, r *CompositionRuntime) error {
	if run.rows == nil {
		return nil
	}
	if len(traces) != len(run.rows) {
		return fmt.Errorf("collection reads differ from actual case count")
	}
	for c := range traces {
		for _, event := range run.rows[c] {
			root := graph.index(r.ReadSites[event.SiteIndex].RootActivityID)
			if root < 0 || root >= len(traces[c].Deliveries) || len(traces[c].Deliveries[root].BlockedBy) != 0 {
				return fmt.Errorf("blocked activity claimed a collection read")
			}
		}
		traces[c].ReadsObserved = true
		traces[c].Reads = run.rows[c]
		traces[c].ReadsDropped, traces[c].ReadRegistrationDropped = run.dropped[c], run.registration[c]
	}
	return graph.bindConstructions(run.construction, traces, r.ReadSites)
}
