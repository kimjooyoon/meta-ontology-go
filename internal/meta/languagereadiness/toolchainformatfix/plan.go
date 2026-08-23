package toolchainformatfix

const (
	unformattedPath = "examples/toolchain-format-fix/unformatted.gooo"
	canonicalPath   = "examples/toolchain-format-fix/canonical.gooo"
	malformedPath   = "examples/toolchain-format-fix/malformed.gooo"
	unformatted     = "package billing\nnamespace billing\nentity Payment id \"billing://entity/payment\"\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Payment\n"
	canonical       = "package billing\nnamespace billing\n\nentity Payment id \"billing://entity/payment\"\nentity Order id \"billing://entity/order\"\n\nactivity PayOrder(Order) -> Payment\n"
)

func argumentsFor(operation string) ([]string, bool) {
	switch operation {
	case "FORMAT_TEXT":
		return []string{"format", unformattedPath}, true
	case "FORMAT_JSON":
		return []string{"format", "--json", unformattedPath}, true
	case "FORMAT_CHECK":
		return []string{"format", "--check", canonicalPath}, true
	case "FIX_JSON_CHANGED":
		return []string{"fix", "--json", unformattedPath}, true
	case "FIX_JSON_FIXED":
		return []string{"fix", "--json", canonicalPath}, true
	case "FIX_TEXT":
		return []string{"fix", unformattedPath}, true
	case "FORMAT_REQUIRED":
		return []string{"format"}, true
	case "FORMAT_MALFORMED":
		return []string{"format", malformedPath}, true
	case "FIX_MALFORMED":
		return []string{"fix", malformedPath}, true
	case "FORMAT_USAGE":
		return []string{"format", "--unknown", canonicalPath}, true
	case "FIX_USAGE":
		return []string{"fix", "--unknown", canonicalPath}, true
	case "FIX_FLAG":
		return []string{"fix", "--write", canonicalPath}, true
	default:
		return nil, false
	}
}
