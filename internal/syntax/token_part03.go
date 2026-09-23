package syntax

import "sort"

var keywordKinds = map[string]TokenKind{
	"package":   TokenPackage,
	"namespace": TokenNamespace,
	"entity":    TokenEntity,
	"id":        TokenID,
	"activity":  TokenActivity,
}

// CanonicalKeywordNames exposes the lexer keyword vocabulary to tools such as
// LSP without allowing callers to mutate the lexer table.
func CanonicalKeywordNames() []string {
	names := make([]string, 0, len(keywordKinds))
	for name := range keywordKinds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
