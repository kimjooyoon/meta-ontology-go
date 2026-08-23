package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	root, input, promotion, guarded, useCases, syntax, diagnostic, packageRuntime          string
	toolchainCLI, toolchainFormatFix, toolchainLSP, toolchainConformance, toolchainRelease string
	output, check, expectedSHA                                                             string
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.root, "root", "", "repository root")
	flag.StringVar(&cfg.input, "input", "", "language concept artifact outside the repository")
	flag.StringVar(&cfg.promotion, "proposal-promotion", "", "verified proposal promotion outside the repository")
	flag.StringVar(&cfg.guarded, "guarded-capability", "", "verified guarded capability outside the repository")
	flag.StringVar(&cfg.useCases, "toolchain-use-cases", "", "verified executable use cases outside the repository")
	flag.StringVar(&cfg.syntax, "language-syntax-roundtrip", "", "verified language syntax receipt outside the repository")
	flag.StringVar(&cfg.diagnostic, "language-diagnostic-provenance", "", "verified diagnostic provenance receipt outside the repository")
	flag.StringVar(&cfg.packageRuntime, "language-package-runtime", "", "verified package runtime receipt outside the repository")
	flag.StringVar(&cfg.toolchainCLI, "toolchain-cli", "", "verified toolchain CLI receipt outside the repository")
	flag.StringVar(&cfg.toolchainFormatFix, "toolchain-format-fix", "", "verified toolchain format/fix receipt outside the repository")
	flag.StringVar(&cfg.toolchainLSP, "toolchain-lsp", "", "verified toolchain LSP receipt outside the repository")
	flag.StringVar(&cfg.toolchainConformance, "toolchain-conformance", "", "verified toolchain conformance receipt outside the repository")
	flag.StringVar(&cfg.toolchainRelease, "toolchain-cross-platform-release", "", "verified cross-platform release receipt outside the repository")
	flag.StringVar(&cfg.output, "output", "", "readiness artifact path outside the repository")
	flag.StringVar(&cfg.check, "check", "", "existing readiness artifact outside the repository")
	flag.StringVar(&cfg.expectedSHA, "expected-sha", "", "exact 40 character commit sha")
	flag.Parse()
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
