package assuranceeligibility

import "encoding/json"

func set(key string, replacement any) func(map[string]any) {
	return func(value map[string]any) { value[key] = replacement }
}

func rewrite(input Input, name string, change func(map[string]any)) {
	var value map[string]any
	if json.Unmarshal(input.Payloads[name], &value) != nil {
		return
	}
	change(value)
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err == nil {
		input.Payloads[name] = append(encoded, '\n')
	}
}
