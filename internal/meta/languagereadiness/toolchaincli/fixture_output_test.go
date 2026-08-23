package toolchaincli

func fakeOutput(key string) (int, string, string) {
	switch key {
	case "version":
		return 0, "gooo 0.1.0-dev (development)\n", ""
	case "version\x00--json":
		return 0, "{\"schema_version\":\"gooo-version/v1\",\"language\":\"gooo\",\"version\":\"0.1.0-dev\",\"status\":\"development\",\"semantic_ir\":\"v1\",\"semantic_check\":\"v1\",\"graph\":\"v1\",\"fix_plan\":\"v1\"}\n", ""
	case "check\x00" + sourceFixture:
		return 0, "ok: " + sourceFixture + "\n", ""
	case "check\x00--json\x00" + sourceFixture:
		return 0, "{\"schema_version\":\"gooo/diagnostics/v1\",\"command\":\"check\",\"status\":\"ok\",\"file\":\"examples/billing/main.gooo\",\"diagnostics\":[]}\n", ""
	case "roundtrip\x00--json\x00" + sourceFixture:
		return 0, "{\"schema_version\":\"gooo/diagnostics/v1\",\"command\":\"roundtrip\",\"status\":\"ok\",\"file\":\"examples/billing/main.gooo\",\"original_semantic_hash\":\"same\",\"round_tripped_semantic_hash\":\"same\",\"equivalent\":true,\"get_put\":true,\"put_get\":true,\"diagnostics\":[]}\n", ""
	case "check\x00--semantic\x00" + sourceFixture:
		return 0, "ok: " + sourceFixture + "\n", deferredProvenance
	case "":
		return 2, "", rootUsage
	case "unknown-contract":
		return 1, "", "gooo: command \"unknown-contract\" is not implemented yet\n"
	case "version\x00--unknown":
		return 2, "", "usage: gooo version [--json]\n"
	case "check":
		return 2, "", "usage: gooo check [--semantic] [--provenance-store <ledger.jsonl>] [--json] <file.gooo>\n"
	case "roundtrip":
		return 2, "", "usage: gooo roundtrip [--json] <file.gooo>\n"
	default:
		return 2, "", "usage: gooo lsp\n"
	}
}
