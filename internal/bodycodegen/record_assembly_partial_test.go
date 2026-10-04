package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPartialFieldChoicesRoundtripThroughSavedJSON(t *testing.T) {
	for _, model := range []string{"", writeThreeContractModel(t)} {
		g, err := NewTypedPathGenerator(model)
		if err != nil {
			t.Fatal(err)
		}
		for _, budget := range []string{"1", "2", "4"} {
			source := []byte(strings.Replace(string(recordAssemblyFixture(t)), `attempts "8"`, `attempts "`+budget+`"`, 1))
			result, err := g.GenerateSourceAssembly(context.Background(), "partial.gooo", source, "Select")
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			var saved Result
			if err = json.Unmarshal(raw, &saved); err != nil {
				t.Fatal(err)
			}
			if _, err = RealizeSourceAssembly(context.Background(), "partial.gooo", source, saved); err != nil {
				t.Fatal("model", model != "", "budget", budget, err)
			}
		}
	}
}
