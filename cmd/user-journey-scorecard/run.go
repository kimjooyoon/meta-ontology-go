package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/userjourneyscorecard"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.executable == "" || cfg.head == "" || cfg.contract == "" || cfg.upstream == "" || cfg.profile == "" || (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("root, executable, head, contract, upstream, and profile plus exactly one output mode are required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	for _, external := range []string{cfg.executable, cfg.upstream, cfg.profile, target} {
		if inside(cfg.root, external) {
			return fmt.Errorf("runtime artifacts must remain outside the repository")
		}
	}
	report, err := userjourneyscorecard.Evaluate(cfg.root, cfg.executable, cfg.head,
		read(cfg.contract), read(cfg.upstream), read(cfg.profile))
	if err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(report, "", "  ")
	raw = append(raw, '\n')
	if cfg.check != "" {
		if !bytes.Equal(read(cfg.check), raw) {
			return fmt.Errorf("FAIL_CLOSED: scorecard replay mismatch")
		}
		return nil
	}
	return os.WriteFile(cfg.output, raw, 0o644)
}

func read(filename string) []byte {
	value, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return value
}

func inside(root, target string) bool {
	root, _ = filepath.Abs(root)
	target, _ = filepath.Abs(target)
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
