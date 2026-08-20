package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

var buildTargets = map[string]struct{}{
	"aix": {}, "android": {}, "darwin": {}, "dragonfly": {}, "freebsd": {},
	"illumos": {}, "ios": {}, "js": {}, "linux": {}, "netbsd": {},
	"openbsd": {}, "plan9": {}, "solaris": {}, "wasip1": {}, "windows": {},
	"386": {}, "amd64": {}, "arm": {}, "arm64": {}, "loong64": {},
	"mips": {}, "mips64": {}, "mips64le": {}, "mipsle": {}, "ppc64": {},
	"ppc64le": {}, "riscv64": {}, "s390x": {}, "wasm": {},
}

func generatedPath(path string, part int) (string, error) {
	dir, base := filepath.Dir(path), filepath.Base(path)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	stem, suffix := peelSuffix(stem, "_test")
	tokens := strings.Split(stem, "_")
	if len(tokens) > 1 {
		if _, ok := buildTargets[tokens[len(tokens)-1]]; ok {
			constraint := "_" + tokens[len(tokens)-1]
			stem = strings.TrimSuffix(stem, constraint)
			suffix = constraint + suffix
			if prior := lastToken(stem); isOS(prior) && isArch(tokens[len(tokens)-1]) {
				stem = strings.TrimSuffix(stem, "_"+prior)
				suffix = "_" + prior + suffix
			}
		}
	}
	if stem == "" || part < 1 {
		return "", fmt.Errorf("invalid generated identity for %s part %s", path, strconv.Itoa(part))
	}
	return filepath.Join(dir, fmt.Sprintf("%s_part%02d%s%s", stem, part, suffix, ext)), nil
}

func peelSuffix(value, suffix string) (string, string) {
	if before, ok := strings.CutSuffix(value, suffix); ok {
		return before, suffix
	}
	return value, ""
}
