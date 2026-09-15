package analyzer

import (
	"strings"
)

func applyAnnotationKey(result *annotation, key, value string) {
	value = strings.TrimSpace(value)
	switch strings.TrimSpace(key) {
	case "kind":
		setAnnotationKind(result, SymbolKind(value))
	case "namespace":
		result.namespace = value
	case "id", "identity":
		if result.id != "" && result.id != value {
			result.conflict = true
		}
		result.id = value
	}
}
func setAnnotationKind(result *annotation, kind SymbolKind) {
	if result.kind != "" && result.kind != kind {
		result.conflict = true
	}
	result.kind = kind
}
func mergeAnnotations(left, right annotation) annotation {
	result := left
	result.active = left.active || right.active
	result.conflict = left.conflict || right.conflict
	if right.kind != "" {
		setAnnotationKind(&result, right.kind)
	}
	if right.namespace != "" {
		result.namespace = right.namespace
	}
	if right.id != "" {
		if result.id != "" && result.id != right.id {
			result.conflict = true
		}
		result.id = right.id
	}
	return result
}
