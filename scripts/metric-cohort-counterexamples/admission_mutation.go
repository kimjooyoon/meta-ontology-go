package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func observedAdmissionVariant(baseline variant, f fixture, planned countercase) (variant, error) {
	if baseline.Inputs.Observation == nil {
		return variant{}, fmt.Errorf("observed admission controls require their baseline inputs")
	}
	v := baseline
	observation := *baseline.Inputs.Observation
	v.Inputs.Observation = &observation
	var err error
	switch planned.Kind {
	case "OBSERVATION_DOCUMENT":
		v.Inputs.Observation.Document, err = admissionDocument(observation.Document, planned.ID)
	case "OBSERVED_RECEIPT":
		v.Receipt, err = admissionReceipt(baseline.Receipt, planned.ID)
	case "POLICY_BYTES":
		v.Inputs.Observation.PolicySource = append([]byte("// admission policy bytes changed\n"), observation.PolicySource...)
	case "DOCUMENT_BYTES":
		v.Inputs.Observation.Document = append(append([]byte(nil), observation.Document...), ' ')
	case "MISSING_OBSERVATION_INPUTS":
		v.Inputs.Observation = nil
	case "LEGACY_DOWNGRADE":
		v.Receipt = f.Raw
	default:
		err = fmt.Errorf("undeclared observed admission kind: %s", planned.Kind)
	}
	return v, err
}

func admissionDocument(raw []byte, id string) ([]byte, error) {
	var document co.ObservationDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if len(document.SubjectSHA) != 40 || len(document.Observations) == 0 || document.Observations[0].Binding == nil {
		return nil, fmt.Errorf("observed admission document needs one source-bound row")
	}
	switch id {
	case "wrong-head":
		document.SubjectSHA = otherSubject(document.SubjectSHA)
	case "wrong-program":
		document.ProgramDigest = artifact.ContentDigest([]byte("admission-other-program"))
	case "wrong-cohort":
		document.DeclaredCohortDigest = artifact.ContentDigest([]byte("admission-other-cohort"))
	case "known-binding":
		document.Observations[0].Binding.Actual += "-admission-changed"
	case "outside-binding":
		document.Observations[0].Operation = outsideObservedOperation1889
	default:
		return nil, fmt.Errorf("undeclared observed admission document: %s", id)
	}
	return json.Marshal(document)
}

func admissionReceipt(raw []byte, id string) ([]byte, error) {
	var receipt co.ObservedReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return nil, err
	}
	switch id {
	case "repository-write-authority":
		receipt.RepositoryWorkspaceWrites = true
	case "promotion-authority":
		receipt.PromotionAuthorized = true
	case "wrong-digest":
		receipt.Digest = artifact.ContentDigest([]byte("admission-forged-digest"))
	case "resealed-source-value":
		old := []byte(`"readiness_credit":false`)
		if bytes.Count(receipt.Decision.Value, old) != 1 {
			return nil, fmt.Errorf("observed admission source value lacks one false credit claim")
		}
		receipt.Decision.Value = bytes.Replace(receipt.Decision.Value, old, []byte(`"readiness_credit":true`), 1)
	default:
		return nil, fmt.Errorf("undeclared observed admission receipt: %s", id)
	}
	if id != "wrong-digest" {
		receipt.Digest = ""
		digest, err := artifact.Digest(receipt)
		if err != nil {
			return nil, err
		}
		receipt.Digest = digest
	}
	return json.Marshal(receipt)
}
