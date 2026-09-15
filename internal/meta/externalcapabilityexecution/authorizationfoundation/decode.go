package authorizationfoundation

import "encoding/json"

func decodeInput(input Input) (Foundation, ArtifactMetadata, Receipt, Receipt, error) {
	foundation, metadata := Foundation{}, ArtifactMetadata{}
	prior, current := Receipt{}, Receipt{}
	values := []struct {
		label string
		raw   []byte
		out   any
	}{
		{"foundation", input.FoundationRaw, &foundation},
		{"artifact-metadata", input.MetadataRaw, &metadata},
		{"prior-receipt", input.PriorReceiptRaw, &prior},
		{"current-receipt", input.CurrentRaw, &current},
	}
	for _, value := range values {
		if len(value.raw) == 0 {
			return foundation, metadata, prior, current, unknown(value.label + " unavailable")
		}
		if err := json.Unmarshal(value.raw, value.out); err != nil {
			return foundation, metadata, prior, current, denied(value.label + " malformed")
		}
	}
	return foundation, metadata, prior, current, nil
}
