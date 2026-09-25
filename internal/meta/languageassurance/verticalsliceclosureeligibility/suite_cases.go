package verticalsliceclosureeligibility

import "encoding/json"

func CaseInput(id, subjectSHA string) (Input, bool) {
	input := EmbeddedInput(subjectSHA)
	switch id {
	case "exact":
	case "assurance-unavailable":
		input.Assurance.Payload = nil
	case "shadow-unavailable":
		input.Shadow.Payload = nil
	case "assurance-digest-mismatch":
		input.Assurance.Payload = append(input.Assurance.Payload, byte(10))
	case "shadow-digest-mismatch":
		input.Shadow.Payload = append(input.Shadow.Payload, byte(10))
	case "unknown-top-decision":
		rewriteShadow(&input, func(value *shadowCapsule) { value.Decision = "UNKNOWN" })
	case "semantic-link-mismatch":
		rewriteShadow(&input, func(value *shadowCapsule) { value.AssuranceDigest = "sha256:unknown" })
	case "observed-write":
		rewriteShadow(&input, func(value *shadowCapsule) { value.RepositoryWrites = 1 })
	default:
		return Input{}, false
	}
	return input, true
}

func rewriteShadow(input *Input, change func(*shadowCapsule)) {
	var value shadowCapsule
	_ = json.Unmarshal(input.Shadow.Payload, &value)
	change(&value)
	input.Shadow.Payload, _ = json.MarshalIndent(value, "", "  ")
}
