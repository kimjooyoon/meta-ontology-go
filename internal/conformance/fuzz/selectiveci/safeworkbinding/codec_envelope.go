package safeworkbinding

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

var bindingFieldOrder = [...]string{
	"schema",
	"task_id",
	"path_id",
	"obligation_id",
	"source_snapshot_digest",
	"semantic_snapshot_digest",
	"policy_digest",
	"registry_digest",
	"toolchain_options_digest",
	"binding_digest",
}

func validateEnvelope(value jsonValue) (SafeWorkBinding, Reason) {
	if value.kind != jsonObjectValue || value.object == nil {
		return SafeWorkBinding{}, ReasonInvalidSchema
	}
	for key := range value.object {
		known := false
		for _, field := range bindingFieldOrder {
			if key == field {
				known = true
				break
			}
		}
		if !known {
			return SafeWorkBinding{}, ReasonUnknownField
		}
	}
	var binding SafeWorkBinding
	for _, field := range bindingFieldOrder {
		fieldValue, present := value.object[field]
		if !present {
			return SafeWorkBinding{}, ReasonRequiredInputMissing
		}
		text, reason := readString(fieldValue)
		if reason != ReasonNone {
			return SafeWorkBinding{}, reason
		}
		switch field {
		case "schema":
			if text != SafeWorkBindingSchemaV1 {
				return SafeWorkBinding{}, ReasonInvalidSchema
			}
			binding.Schema = text
		case "task_id", "path_id", "obligation_id":
			if !validateStableID(text) {
				return SafeWorkBinding{}, ReasonInvalidStableID
			}
			switch field {
			case "task_id":
				binding.TaskID = StableID(text)
			case "path_id":
				binding.PathID = StableID(text)
			case "obligation_id":
				binding.ObligationID = StableID(text)
			}
		case "source_snapshot_digest", "semantic_snapshot_digest", "policy_digest",
			"registry_digest", "toolchain_options_digest", "binding_digest":
			if !validateDigest(text) {
				return SafeWorkBinding{}, ReasonInvalidDigest
			}
			switch field {
			case "source_snapshot_digest":
				binding.SourceSnapshotDigest = Digest(text)
			case "semantic_snapshot_digest":
				binding.SemanticSnapshotDigest = Digest(text)
			case "policy_digest":
				binding.PolicyDigest = Digest(text)
			case "registry_digest":
				binding.RegistryDigest = Digest(text)
			case "toolchain_options_digest":
				binding.ToolchainOptionsDigest = Digest(text)
			case "binding_digest":
				binding.BindingDigest = Digest(text)
			}
		}
	}
	if binding.BindingDigest != bindingDigest(binding) {
		return SafeWorkBinding{}, ReasonBindingDigestMismatch
	}
	return binding, ReasonNone
}

func readString(value jsonValue) (string, Reason) {
	switch value.kind {
	case jsonNullValue:
		return "", ReasonNullValue
	case jsonStringValue:
		if value.text == "" {
			return "", ReasonEmptyValue
		}
		return value.text, ReasonNone
	default:
		return "", ReasonInvalidSchema
	}
}

func validateStableID(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}
	if value == "" {
		return false
	}
	if len(value) > 256 {
		return false
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return false
	}
	identity, err := semantic.ParseIdentity(value)
	if err != nil {
		return false
	}
	return identity.String() == value
}

func validateDigest(value string) bool {
	payload, found := strings.CutPrefix(value, "sha256:")
	return found && len(payload) == 64 && strings.Trim(payload, "0123456789abcdef") == ""
}
