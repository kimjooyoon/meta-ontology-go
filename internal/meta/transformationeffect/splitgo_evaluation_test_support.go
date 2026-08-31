package transformationeffect

import (
	"encoding/json"
	"reflect"
)

var splitGoTestIndicatorIDs = []string{
	"filesystem.atomic-replacement/v1",
	"go.filename.build-semantics/v1",
	"go.header.preserved/v1",
	"go.import.identity/v1",
	"go.initialization.order/v1",
	"go.package.conformance/v1",
}

func splitGoTestReport(decision, verdict string, ids []string) []byte {
	indicators := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		indicators = append(indicators, map[string]string{
			"indicator_id": id,
			"verdict":      verdict,
		})
	}
	report, err := json.Marshal(map[string]any{
		"decision":   decision,
		"indicators": indicators,
	})
	if err != nil {
		panic(err)
	}
	return report
}

func splitGoReceiptField(receipt any, names ...string) string {
	value := reflect.ValueOf(receipt)
	for _, name := range names {
		field := value.FieldByName(name)
		if field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}
	return ""
}
