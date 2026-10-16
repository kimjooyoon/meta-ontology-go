package main

import (
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
)

func run(cfg config) error {
	started := time.Now()
	digest, err := canonicalDigest(cfg.Origin.ArtifactDigest)
	if err != nil {
		return err
	}
	cfg.Origin.ArtifactDigest = digest
	if err := preparePaths(&cfg); err != nil {
		return err
	}
	planned := countercases()
	result := report{Schema: "gooo/source-cohort-native-counterexamples/v1", Decision: "FAIL_CLOSED",
		Origin: cfg.Origin, RegistryOrigin: metricprogram.RegistrySourceOrigin(), CasesDeclared: len(planned),
		ProducerBinding: "UNASSESSED",
		VariantPolicy:   "ADVERSARIAL_RESEALED_INPUTS_ARE_NOT_UPSTREAM_AUTHORITY",
		UnrelatedScope:  "Existing positive workflow compares23unrelated results; refused cases return no readiness snapshot.",
		ElapsedScope:    "RUN_BEFORE_FINAL_REPORT_ENCODING",
	}
	result.ExecutionOrigin, err = observeExecution(cfg)
	var f fixture
	if err == nil {
		f, err = loadFixture(cfg)
	}
	if err == nil {
		result.ProducerBinding = "EXACT_ORIGINAL_ARCHIVE"
		err = verifyBaseline(cfg, f, &result)
	}
	if err != nil {
		result.SetupError = err.Error()
		result.ElapsedNanoseconds = time.Since(started).Nanoseconds()
		if writeErr := writeReport(cfg.Output, result); writeErr != nil {
			return writeErr
		}
		return err
	}
	result.SetupElapsedNanoseconds = time.Since(started).Nanoseconds()
	for _, row := range planned {
		observed, caseErr := executeCase(cfg, f, row)
		recordCase(&result, observed, caseErr)
	}
	upstream, upstreamErr := observeUnregisteredUpstream(cfg, f)
	result.UnregisteredUpstream = &upstream
	observed, observedErr := observeSeparateObservations(cfg, f)
	result.ObservedCohort = &observed
	admission, admissionErr := observeAdmissionControls(cfg, f)
	result.ObservedAdmission = &admission
	if result.CasesAttempted == result.CasesDeclared && result.CasesPassed == result.CasesDeclared &&
		result.CasesExecuted == result.CasesDeclared && upstreamErr == nil && upstream.Passed &&
		observedErr == nil && observed.Decision == "PASS" && admissionErr == nil && admission.Decision == "PASS" {
		result.Decision = "PASS"
	}
	finalOrigin, observationErr := observeExecution(cfg)
	if observationErr != nil || finalOrigin == nil || result.ExecutionOrigin == nil || *finalOrigin != *result.ExecutionOrigin {
		result.Decision = "FAIL_CLOSED"
		result.SetupError = fmt.Sprintf("countercase execution origin changed: %v", observationErr)
	}
	result.ElapsedNanoseconds = time.Since(started).Nanoseconds()
	if err := writeReport(cfg.Output, result); err != nil {
		return err
	}
	if result.Decision != "PASS" {
		return fmt.Errorf("native counterexamples passed %d/%d; see report.json", result.CasesPassed, result.CasesDeclared)
	}
	return nil
}

func recordCase(result *report, row caseResult, err error) {
	result.Cases = append(result.Cases, row)
	result.CasesAttempted++
	if row.ConsumerAttempted {
		result.CasesExecuted++
	}
	if row.Passed && row.ConsumerAttempted && err == nil {
		result.CasesPassed++
	}
}
