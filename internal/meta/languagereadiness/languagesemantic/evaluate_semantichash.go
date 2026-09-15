package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func semanticHash(value string) string {
	return semantic.StableHashString(value)
}
