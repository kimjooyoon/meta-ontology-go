package formatter

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

func kebab(value string) string {
	var result strings.Builder
	for index, r := range value {
		if unicode.IsUpper(r) && index > 0 {
			result.WriteByte('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// SemanticFingerprint returns a canonical identity-and-relation view. It
// deliberately excludes display names and declaration order.
func (d Document) SemanticFingerprint() string {
	if d.validate().HasErrors() {
		return ""
	}
	entities := make(map[string]string, len(d.Declarations))
	records := make([]string, 0, len(d.Declarations))
	for _, declaration := range d.Declarations {
		if declaration.Kind == EntityDeclaration {
			entities[declaration.Name] = canonicalIdentity(declaration.ID)
			records = append(records, "entity|"+canonicalIdentity(declaration.ID))
		}
	}
	for _, declaration := range d.Declarations {
		if declaration.Kind != ActivityDeclaration {
			continue
		}
		activityID := declaration.ID
		if activityID == "" {
			activityID = defaultActivityID(d.Namespace, declaration.Name)
		}
		activityID = canonicalIdentity(activityID)
		inputs := make([]string, len(declaration.Inputs))
		for index, input := range declaration.Inputs {
			inputs[index] = entities[input]
		}
		output := entities[declaration.Output]
		records = append(records, "activity|"+activityID+"|used="+strings.Join(inputs, ",")+"|generated="+output)
	}
	sort.Strings(records)
	return fmt.Sprintf("package=%s|namespace=%s|%s", d.Package, d.Namespace, strings.Join(records, ";"))
}
