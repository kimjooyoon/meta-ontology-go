package directorypartition

import (
	"bytes"
	"testing"
)

func TestOntologyBindsReportSchema(t *testing.T) {
	ontologyDigest, err := validateOntology()
	if err != nil {
		t.Fatal(err)
	}
	report, err := sealReport(Report{
		Schema: ReportSchema, Repository: "kimjooyoon/meta-ontology-go",
		SubjectSHA: "0123456789abcdef", OntologyDigest: ontologyDigest,
		RootPolicy: RootPolicy{
			CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE",
			TopologyReason: "ROOT_TOPOLOGY_EXEMPT", READMERequirement: "NOT_APPLICABLE",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if report.Digest == "" || !bytes.Contains(payload, []byte(ReportSchema)) {
		t.Fatal("report schema is not sealed by the meta ontology")
	}
}
