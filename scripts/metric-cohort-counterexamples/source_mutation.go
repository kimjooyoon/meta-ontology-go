package main

import (
	"encoding/json"
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
)

func sourceMutation(inputs co.SourceInputs, id string) (co.SourceInputs, error) {
	if id == "program-source-tamper" {
		inputs.ProgramSource = append(append([]byte(nil), inputs.ProgramSource...), []byte("\n// native tamper\n")...)
		return inputs, nil
	}
	var program metricprogram.Program
	if err := json.Unmarshal(inputs.Program, &program); err != nil {
		return inputs, err
	}
	switch id {
	case "resealed-program":
		program.SourceDigest = artifact.ContentDigest([]byte("native-resealed-program-source"))
	case "wrong-source-subject":
		program.SubjectSHA = otherSubject(program.SubjectSHA)
	default:
		return inputs, fmt.Errorf("undeclared source mutation: %s", id)
	}
	program.Digest = ""
	digest, err := artifact.Digest(program)
	if err != nil {
		return inputs, err
	}
	program.Digest = digest
	inputs.Program, err = json.Marshal(program)
	if err != nil || id != "resealed-program" {
		return inputs, err
	}
	var verification programverify.Report
	if err := json.Unmarshal(inputs.ProgramVerification, &verification); err != nil {
		return inputs, err
	}
	verification.ProgramDigest, verification.SourceDigest = program.Digest, program.SourceDigest
	verification.Digest = ""
	digest, err = artifact.Digest(verification)
	if err != nil {
		return inputs, err
	}
	verification.Digest = digest
	inputs.ProgramVerification, err = json.Marshal(verification)
	return inputs, err
}
