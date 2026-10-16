package main

import (
	"encoding/json"
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func mutatedReceipt(f fixture, id string) ([]byte, error) {
	receipt := f.Receipt
	receipt.Expected = append([]co.OperationBinding(nil), receipt.Expected...)
	switch id {
	case "omitted-declared-subject":
		if len(receipt.Expected) < 2 {
			return nil, fmt.Errorf("declared omission case needs at least two expected subjects")
		}
		receipt.Expected = receipt.Expected[:len(receipt.Expected)-1]
		receipt.ExpectedCount, receipt.BoundCount = len(receipt.Expected), len(receipt.Expected)
	case "replaced-declared-subject":
		receipt.Expected[0].Operation = "native-replaced-declared-subject"
		receipt.Expected[0].IndicatorID = metricstrategy.ConceptOperationIndicatorID(receipt.Expected[0].Operation)
	case "wrong-head":
		receipt.SubjectSHA = otherSubject(receipt.SubjectSHA)
	case "wrong-source-digest":
		receipt.ProgramSourceDigest = artifact.ContentDigest([]byte("native-wrong-source"))
	case "repository-write-authority":
		receipt.RepositoryWorkspaceWrites = true
	case "promotion-authority":
		receipt.PromotionAuthorized = true
	case "wrong-registry-origin":
		receipt.RegistryOrigin.ProjectionDigest = artifact.ContentDigest([]byte("native-wrong-registry"))
	default:
		return nil, fmt.Errorf("undeclared receipt mutation: %s", id)
	}
	receipt.Digest = ""
	digest, err := artifact.Digest(receipt)
	if err != nil {
		return nil, err
	}
	receipt.Digest = digest
	return json.Marshal(receipt)
}

func otherSubject(subject string) string {
	if subject[0] == '0' {
		return "1" + subject[1:]
	}
	return "0" + subject[1:]
}
