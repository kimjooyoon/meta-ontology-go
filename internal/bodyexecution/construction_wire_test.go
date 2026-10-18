package bodyexecution

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestConstructionOriginFlatWireCompatibility(t *testing.T) {
	for _, kind := range []string{"", "map_values"} {
		origin := bodycodegen.BodyConstructionOrigin{ActivityID: "urn:root", BodySHA256: "sha256:body", Start: 7, End: 12, Expression: "items", Elements: []bodycodegen.BodyElementOrigin{}, Kind: kind}
		if kind != "" {
			origin.MapperActivityID = "urn:mapper"
		}
		raw, _ := json.Marshal(origin)
		var saved bodycodegen.BodyConstructionOrigin
		if err := DecodeExecutionReceipt(raw, &saved); err != nil || !reflect.DeepEqual(origin, saved) {
			t.Fatal("flat source coordinates", err, string(raw))
		}
		for _, bad := range []string{strings.Replace(string(raw), `"decoded_start":7`, `"DecodedStart":7`, 1), strings.Replace(string(raw), `"decoded_start":7`, `"decoded_start":7,"decoded_start":7`, 1)} {
			if DecodeExecutionReceipt([]byte(bad), &saved) == nil {
				t.Fatal("ambiguous coordinate accepted", bad)
			}
		}
	}
}
