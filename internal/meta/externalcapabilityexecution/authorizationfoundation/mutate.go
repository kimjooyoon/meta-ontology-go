package authorizationfoundation

import "encoding/json"

func mutateObject(raw []byte, mutate func(map[string]any)) []byte {
	value := map[string]any{}
	if json.Unmarshal(raw, &value) != nil {
		return raw
	}
	mutate(value)
	result, _ := json.Marshal(value)
	return result
}

func suiteInput(input Input, kind string) Input {
	result := input
	switch kind {
	case "metadata-unavailable":
		result.MetadataRaw = nil
	case "prior-unavailable":
		result.PriorReceiptRaw = nil
	case "current-unavailable":
		result.CurrentRaw = nil
	case "metadata-expired":
		result.MetadataRaw = mutateObject(result.MetadataRaw, func(value map[string]any) { value["expired"] = true })
	case "artifact-id-mismatch":
		result.FoundationRaw = mutateObject(result.FoundationRaw, func(value map[string]any) { value["artifact_id"] = 1 })
	case "archive-digest-mismatch":
		result.MetadataRaw = mutateObject(result.MetadataRaw, func(value map[string]any) { value["digest"] = "sha256:bad" })
	case "prior-receipt-tamper":
		result.PriorReceiptRaw = append(append([]byte(nil), result.PriorReceiptRaw...), '\n')
	case "source-digest-mismatch":
		result.CurrentRaw = mutateObject(result.CurrentRaw, func(value map[string]any) { value["policy_source_digest"] = "sha256:bad" })
	case "tree-digest-mismatch":
		result.CurrentRaw = mutateObject(result.CurrentRaw, func(value map[string]any) { value["policy_generated_digest"] = "sha256:bad" })
	case "unknown-reason-mismatch":
		result.CurrentRaw = mutateObject(result.CurrentRaw, func(value map[string]any) { value["unknowns"] = []any{} })
	case "authority-ceiling":
		result.CurrentRaw = mutateObject(result.CurrentRaw, func(value map[string]any) { value["execution_authority"] = true })
	}
	return result
}
