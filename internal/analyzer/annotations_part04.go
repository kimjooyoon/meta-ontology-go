package analyzer

import (
	"go/ast"
	"strings"
)

func namedSpec(specification ast.Spec, name string) bool {
	switch current := specification.(type) {
	case *ast.TypeSpec:
		return current.Name.Name == name
	case *ast.ValueSpec:
		for _, candidate := range current.Names {
			if candidate.Name == name {
				return true
			}
		}
	}
	return false
}
func parseAnnotations(group *ast.CommentGroup) annotation {
	var result annotation
	if group == nil {
		return result
	}
	for _, comment := range group.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimSpace(strings.TrimSuffix(text, "*/"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "//"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "/*"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "*"))
		if !strings.HasPrefix(text, "gooo:") {
			continue
		}
		parts := annotationFields(strings.TrimSpace(strings.TrimPrefix(text, "gooo:")))
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "semantic":
			result.active = true
			applyAnnotationParts(&result, parts[1:])
		case "activity", "entity":
			result.active = true
			setAnnotationKind(&result, SymbolKind(parts[0]))
			applyAnnotationParts(&result, parts[1:])
		case "generated:start", "generated:end", "slot:start", "slot:end":
			continue
		default:
			applyAnnotationParts(&result, parts)
		}
	}
	return result
}
func applyAnnotationParts(result *annotation, parts []string) {
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if key, value, ok := strings.Cut(part, "="); ok {
			applyAnnotationKey(result, key, value)
			continue
		}
		switch part {
		case string(KindActivity), string(KindEntity):
			setAnnotationKind(result, SymbolKind(part))
		default:
			if result.id == "" {
				result.id = part
			}
		}
	}
}
