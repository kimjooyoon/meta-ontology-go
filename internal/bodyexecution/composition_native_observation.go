package bodyexecution

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type compositionNativeObservation struct {
	first          []byte
	conditions     []byte
	conditionRows  [][]CompositionConditionObservation
	conditionDrops []int
	reads          compositionReadRun
	locals         compositionLocalRun
}

func (run *compositionNativeObservation) reset(root string, r *CompositionRuntime) error {
	if err := run.reads.reset(root, r); err != nil {
		return err
	}
	if err := run.locals.reset(root, r); err != nil {
		return err
	}
	if len(r.ConditionSites) == 0 {
		return nil
	}
	err := os.Remove(filepath.Join(root, conditionObservationFile))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (run *compositionNativeObservation) observeConditions(attempt int, root string, graph compositionGraph, suite CompositionCases, r *CompositionRuntime) error {
	if len(r.ConditionSites) == 0 {
		return nil
	}
	raw, err := readNativeConditionFile(filepath.Join(root, conditionObservationFile))
	if err != nil {
		return err
	}
	if attempt != 0 {
		if !bytes.Equal(run.conditions, raw) {
			return fmt.Errorf("compiled native if observation files differ")
		}
		return nil
	}
	run.conditions = raw
	run.conditionRows, run.conditionDrops, r.ConditionObservation, err = graph.nativeConditionFile(raw, len(suite.Cases), r.ConditionSites)
	return err
}

func (run *compositionNativeObservation) observe(attempt int, root string, output []byte, graph compositionGraph, suite CompositionCases, r *CompositionRuntime) error {
	if err := run.reads.observe(attempt, root, graph, suite, r); err != nil {
		return err
	}
	if err := run.locals.observe(attempt, root, r); err != nil {
		return err
	}
	if err := run.observeConditions(attempt, root, graph, suite, r); err != nil {
		return err
	}
	if attempt != 0 {
		if !bytes.Equal(run.first, output) {
			return fmt.Errorf("compiled composition replay outputs differ")
		}
		return nil
	}
	run.first = output
	traces, passed, outcomes, err := graph.nativeOutcomeTraces(output, suite, r.FaultSites)
	if err != nil {
		return err
	}
	if err := graph.bindNativeConditions(traces, run.conditionRows, r.ConditionSites); err != nil {
		return err
	}
	for c, lost := range run.conditionDrops {
		traces[c].ConditionsDropped = lost
	}
	if err := run.reads.bind(graph, traces, r); err != nil {
		return err
	}
	if err := run.locals.bind(graph, traces, r); err != nil {
		return err
	}
	r.Traces, r.FinitePassed, r.Outcomes = traces, passed, outcomes
	return nil
}
