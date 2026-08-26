package toolchainformatfix

func fakeGuardrail(key string) (int, string, string) {
	switch key {
	case "format":
		return 2, "", "usage: gooo format [--check] [--json] <file.gooo>\n"
	case "format\x00" + malformedPath:
		return 1, "", "gooo: " + malformedPath + ": FORMAT_FIX_SOURCE_UNKNOWN\n"
	case "fix\x00" + malformedPath:
		return 1, "", "gooo: " + malformedPath + ": FORMAT_FIX_SOURCE_UNKNOWN\n"
	case "format\x00--unknown\x00" + canonicalPath:
		return 2, "", "usage: gooo format [--check] [--json] <file.gooo>\n"
	default:
		return 2, "", "usage: gooo fix [--json] <file.gooo>\n"
	}
}
