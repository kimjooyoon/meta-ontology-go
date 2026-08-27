package syntax

var keywordKinds = map[string]TokenKind{
	"package":   TokenPackage,
	"namespace": TokenNamespace,
	"entity":    TokenEntity,
	"id":        TokenID,
	"activity":  TokenActivity,
	"freshness": TokenFreshness,
}
