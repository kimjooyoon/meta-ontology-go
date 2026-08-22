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
	if err := requireExternal(cfg.root, cfg.input, target); err != nil {
		return err
	}
	if cfg.check != "" {
		return consume(cfg, stdout)
	}
	return produce(cfg, stdout)
}
