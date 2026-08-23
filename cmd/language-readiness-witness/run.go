package main

import (
	"fmt"
	"io"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.input == "" || cfg.expectedSHA == "" {
		return fmt.Errorf("root, input, and expected-sha are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	paths := completePaths(cfg, target)
	if (cfg.guarded == "") != (cfg.useCases == "") ||
		(cfg.guarded == "") != (cfg.syntax == "") ||
		(cfg.guarded == "") != (cfg.diagnostic == "") ||
		(cfg.guarded == "") != (cfg.packageRuntime == "") ||
		(cfg.guarded == "") != (cfg.toolchainCLI == "") ||
		(cfg.guarded == "") != (cfg.toolchainFormatFix == "") {
		return fmt.Errorf("guarded-capability, use-cases, syntax, diagnostic, package-runtime, toolchain-cli, and toolchain-format-fix evidence must be provided together")
	}
	if cfg.toolchainConformance != "" && cfg.guarded == "" {
		return fmt.Errorf("toolchain-conformance requires the complete evidence set")
	}
	if cfg.toolchainLSP != "" && cfg.toolchainConformance == "" {
		return fmt.Errorf("toolchain-lsp requires toolchain-conformance evidence")
	}
	if cfg.toolchainRelease != "" && cfg.toolchainLSP == "" {
		return fmt.Errorf("toolchain cross-platform release requires toolchain-lsp evidence")
	}
	if err := requireExternal(cfg.root, paths...); err != nil {
		return err
	}
	if cfg.check != "" {
		return consume(cfg, stdout)
	}
	return produce(cfg, stdout)
}
