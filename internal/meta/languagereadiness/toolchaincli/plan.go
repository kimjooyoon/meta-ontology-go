package toolchaincli

const sourceFixture = "examples/billing/main.gooo"

func argumentsFor(operation string) ([]string, bool) {
	switch operation {
	case "VERSION_TEXT":
		return []string{"version"}, true
	case "VERSION_JSON":
		return []string{"version", "--json"}, true
	case "CHECK_TEXT":
		return []string{"check", sourceFixture}, true
	case "CHECK_JSON":
		return []string{"check", "--json", sourceFixture}, true
	case "ROUNDTRIP_JSON":
		return []string{"roundtrip", "--json", sourceFixture}, true
	case "SEMANTIC_CHECK":
		return []string{"check", "--semantic", sourceFixture}, true
	case "ROOT_USAGE":
		return []string{}, true
	case "UNKNOWN_COMMAND":
		return []string{"unknown-contract"}, true
	case "VERSION_USAGE":
		return []string{"version", "--unknown"}, true
	case "CHECK_USAGE":
		return []string{"check"}, true
	case "ROUNDTRIP_USAGE":
		return []string{"roundtrip"}, true
	case "LSP_USAGE":
		return []string{"lsp", "unexpected"}, true
	default:
		return nil, false
	}
}
