package conceptoperation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
)

func TestReceiptRegistryOriginFresh1639(t *testing.T) {
	for _, field := range []string{"source", "projection", "operation-count", "entity-count"} {
		t.Run(field, func(t *testing.T) {
			origin := metricprogram.RegistrySourceOrigin()
			switch field {
			case "source":
				origin.SourceDigest = "sha256:" + strings.Repeat("1", 64)
			case "projection":
				origin.ProjectionPath = "another/projection.go"
			case "operation-count":
				origin.OperationCount--
			case "entity-count":
				origin.EntityCount--
			}
			receipt, err := seal(Receipt{RegistryOrigin: origin})
			if err != nil {
				t.Fatal(err)
			}
			if err := VerifyReceipt(receipt, "", ""); err == nil ||
				!strings.Contains(err.Error(), "registry origin") {
				t.Fatalf("resealed changed registry origin: %v", err)
			}
		})
	}
}
