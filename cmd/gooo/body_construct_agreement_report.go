package main

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type constructionAgreementState uint8

const (
	constructionAgreementUnknown constructionAgreementState = iota
	constructionAgreementPassed
	constructionAgreementFailed
)

// Compare recorded finite obligations without changing selection or judging
// untried implementations. An unavailable side cannot become a failed side.
func constructionAgreementRows(c bodyexecution.JointConstruction) [][2]string {
	if len(c.Attempts) == 0 {
		return nil
	}
	both, localOnly, callerOnly, neither, unknown := 0, 0, 0, 0, 0
	localWitness, callerWitness := -1, -1
	for i, attempt := range c.Attempts {
		local, caller := constructionLocalAgreement(attempt), constructionCallerAgreement(attempt.Runtime)
		switch {
		case local == constructionAgreementUnknown || caller == constructionAgreementUnknown:
			unknown++
		case local == constructionAgreementPassed && caller == constructionAgreementPassed:
			both++
		case local == constructionAgreementPassed:
			localOnly++
			if localWitness < 0 {
				localWitness = i
			}
		case caller == constructionAgreementPassed:
			callerOnly++
			if callerWitness < 0 {
				callerWitness = i
			}
		default:
			neither++
		}
	}
	rows := [][2]string{{"Local/caller agreement across attempts", fmt.Sprintf(
		"%d both satisfied; %d local only; %d caller only; %d neither; %d unclassified (missing, unavailable or inconsistent observations)",
		both, localOnly, callerOnly, neither, unknown)}}
	if localWitness >= 0 {
		rows = append(rows, [2]string{"Local-only witness", constructionCallerWitness(localWitness, c.Attempts[localWitness])})
	}
	if callerWitness >= 0 {
		rows = append(rows, [2]string{"Caller-only witness", constructionLocalWitness(callerWitness, c.Attempts[callerWitness])})
	}
	return append(rows, [2]string{"Agreement scope", "fully observed local and caller checks for recorded attempts; untried choices and other inputs remain unassessed"})
}

func constructionLocalAgreement(a bodyexecution.JointAttempt) constructionAgreementState {
	if a.Rejection != nil || a.LocalTotal < 0 || a.LocalPassed < 0 || a.LocalPassed > a.LocalTotal ||
		a.CriteriaTotal < 0 || a.CriteriaPassed < 0 || a.CriteriaFailed < 0 || a.CriteriaUnknown < 0 ||
		a.CriteriaPassed > a.CriteriaTotal || a.CriteriaFailed > a.CriteriaTotal-a.CriteriaPassed ||
		a.CriteriaUnknown != a.CriteriaTotal-a.CriteriaPassed-a.CriteriaFailed || a.CriteriaUnknown != 0 ||
		(a.LocalTotal == 0 && a.CriteriaTotal == 0) {
		return constructionAgreementUnknown
	}
	if a.LocalPassed == a.LocalTotal && a.CriteriaPassed == a.CriteriaTotal {
		return constructionAgreementPassed
	}
	return constructionAgreementFailed
}

func constructionCallerAgreement(r bodyexecution.CompositionRuntime) constructionAgreementState {
	matched, mismatched, faulted, blocked, extraFaults := constructionRuntimeCounts(r)
	if r.Stage != "COMPLETE" || r.Failure != "" || r.FiniteTotal <= 0 ||
		matched != r.FinitePassed || matched+mismatched != r.FiniteTotal ||
		faulted != 0 || blocked != 0 || extraFaults != 0 {
		return constructionAgreementUnknown
	}
	if mismatched == 0 {
		return constructionAgreementPassed
	}
	return constructionAgreementFailed
}

func constructionCallerWitness(index int, a bodyexecution.JointAttempt) string {
	for _, trace := range a.Runtime.Traces {
		for _, d := range trace.Deliveries {
			if len(d.Expected) != 0 && d.Passed != nil && !*d.Passed && d.Fault == nil && len(d.BlockedBy) == 0 {
				return fmt.Sprintf("attempt %d satisfies local checks; %s", index+1, constructionDeliveryDetail(trace.CaseIndex, d))
			}
		}
	}
	return fmt.Sprintf("attempt %d satisfies local checks; inspect recorded caller outputs", index+1)
}

func constructionLocalWitness(index int, a bodyexecution.JointAttempt) string {
	for _, candidate := range a.Candidates {
		e := candidate.Attempt.Criteria
		if e == nil {
			continue
		}
		for _, input := range e.Inputs {
			for _, check := range input.Checks {
				if check.Status == "UNSATISFIED" {
					return fmt.Sprintf("attempt %d matches caller outputs; %s (%s), criterion %s returned false for input %s; actual %s",
						index+1, e.Activity, e.ActivityID, check.ID,
						constructionValue(json.RawMessage(input.InputJSON)), constructionValue(json.RawMessage(input.ActualJSON)))
				}
			}
		}
	}
	return fmt.Sprintf("attempt %d matches caller outputs; local checks %d/%d and criteria %d/%d satisfied; inspect JSON for values",
		index+1, a.LocalPassed, a.LocalTotal, a.CriteriaPassed, a.CriteriaTotal)
}
