package main

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionRuntimeShare(r bodyexecution.CompositionRuntime) string {
	if r.FiniteTotal <= 0 {
		return "not measured (no scored expectations)"
	}
	matched, mismatched, faulted, blocked, extraFaults := constructionRuntimeCounts(r)
	observed := matched + mismatched + faulted + blocked
	if observed > r.FiniteTotal || matched != r.FinitePassed {
		return "inconsistent recorded counts; inspect JSON"
	}
	if observed == 0 {
		return fmt.Sprintf("not measured (%d expected outputs; 0 observed)", r.FiniteTotal)
	}
	return fmt.Sprintf("%s matched; %d mismatched, %d faulted, %d blocked, %d unobserved; %d additional unscored faults/blocks",
		constructionShare(matched, r.FiniteTotal), mismatched, faulted, blocked, r.FiniteTotal-observed, extraFaults)
}

func constructionRuntimeCounts(r bodyexecution.CompositionRuntime) (matched, mismatched, faulted, blocked, extraFaults int) {
	for _, trace := range r.Traces {
		for _, d := range trace.Deliveries {
			if len(d.Expected) == 0 {
				if d.Fault != nil || len(d.BlockedBy) > 0 {
					extraFaults++
				}
				continue
			}
			switch {
			case d.Fault != nil:
				faulted++
			case len(d.BlockedBy) > 0:
				blocked++
			case d.Passed != nil && *d.Passed:
				matched++
			case d.Passed != nil:
				mismatched++
			}
		}
	}
	return
}

func constructionFailureRows(result bodyConstructOutput) [][2]string {
	rows := [][2]string{}
	if failure := result.Construction.Failure; failure != "" {
		rows = append(rows, [2]string{"Construction failure", failure})
	}
	if failure := result.Evaluation.Runtime.Failure; failure != "" {
		rows = append(rows, [2]string{"Evaluation failure", failure})
	}
	if replay := result.Evaluation.ReplayFailure; replay != nil {
		rows = append(rows, [2]string{"Saved replay failure", fmt.Sprintf("attempt %d, stage %s: %s", replay.AttemptIndex+1, replay.Stage, replay.Runtime.Failure)})
	}
	count := 0
	for _, trace := range result.Evaluation.Runtime.Traces {
		for _, d := range trace.Deliveries {
			if d.Fault == nil && len(d.BlockedBy) == 0 && (d.Passed == nil || *d.Passed) {
				continue
			}
			count++
			if count <= 8 {
				rows = append(rows, [2]string{"Evaluation observation", constructionDeliveryDetail(trace.CaseIndex, d)})
			}
		}
	}
	if count > 8 {
		rows = append(rows, [2]string{"Additional failed/faulted/blocked outputs", fmt.Sprintf("%d; use --format json", count-8)})
	}
	return rows
}

func constructionDeliveryDetail(index int, d bodyexecution.CompositionDelivery) string {
	reason := "value mismatch"
	if d.Fault != nil {
		reason = "fault: " + d.Fault.Kind
	} else if len(d.BlockedBy) > 0 {
		reason = "blocked by unavailable producer"
	}
	return fmt.Sprintf("case index %d, activity %s: %s; actual %s; expected %s", index,
		d.ActivityID, reason, constructionValue(d.Actual), constructionValue(d.Expected))
}

func constructionValue(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "not recorded"
	}
	value, err := json.Marshal(raw)
	if err != nil {
		return "invalid recorded JSON"
	}
	text := []rune(string(value))
	if len(text) > 120 {
		return string(text[:120]) + "… (full value in --format json)"
	}
	return string(text)
}
