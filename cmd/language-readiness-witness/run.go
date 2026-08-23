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
	paths := []string{cfg.input, target}
	if cfg.promotion != "" {
		paths = append(paths, cfg.promotion)
	}
	if cfg.guarded != "" {
		paths = append(paths, cfg.guarded)
	}
	if cfg.useCases != "" {
		paths = append(paths, cfg.useCases)
	}
	if cfg.syntax != "" {
		paths = append(paths, cfg.syntax)
	}
	if cfg.diagnostic != "" {
		paths = append(paths, cfg.diagnostic)
	}
	if cfg.packageRuntime != "" {
		paths = append(paths, cfg.packageRuntime)
	}
	if cfg.toolchainCLI != "" {
		paths = append(paths, cfg.toolchainCLI)
	}
	if cfg.toolchainFormatFix != "" {
		paths = append(paths, cfg.toolchainFormatFix)
	}
	if (cfg.guarded == "") != (cfg.useCases == "") ||
		(cfg.guarded == "") != (cfg.syntax == "") ||
		(cfg.guarded == "") != (cfg.diagnostic == "") ||
		(cfg.guarded == "") != (cfg.packageRuntime == "") ||
		(cfg.guarded == "") != (cfg.toolchainCLI == "") ||
		(cfg.guarded == "") != (cfg.toolchainFormatFix == "") {
		return fmt.Errorf("guarded-capability, use-cases, syntax, diagnostic, package-runtime, toolchain-cli, and toolchain-format-fix evidence must be provided together")
	}
	if err := requireExternal(cfg.root, paths...); err != nil {
		return err
	}
	if cfg.check != "" {
		return consume(cfg, stdout)
	}
	return produce(cfg, stdout)
}
