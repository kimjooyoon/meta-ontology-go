package verify

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
)

func TestRegistryMeaningFresh1633(t *testing.T) {
	t.Run("operation-meaning", func(t *testing.T) {
		actual, err := json.Marshal(metricprogram.CanonicalOperations())
		if err != nil {
			t.Fatal(err)
		}
		expected, err := json.Marshal(operations)
		if err != nil || !bytes.Equal(actual, expected) {
			t.Fatal("Gooo projection differs from the independent nine-operation compatibility kernel")
		}
	})
	t.Run("canonical-program", func(t *testing.T) {
		if !bytes.Equal(metricprogram.CanonicalSource(), expectedSource()) {
			t.Fatal("Gooo declaration changes the independently reconstructed program source")
		}
	})
	t.Run("declared-denominator", func(t *testing.T) {
		origin := metricprogram.RegistrySourceOrigin()
		if origin.OperationCount != int64(len(operations)) || origin.EntityCount != int64(len(entities)) {
			t.Fatal("declared counts differ from the compatibility kernel")
		}
	})
}
