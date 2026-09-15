package main

import (
	"path/filepath"
	"strings"
)

var operatingSystems = map[string]struct{}{
	"aix": {}, "android": {}, "darwin": {}, "dragonfly": {}, "freebsd": {},
	"illumos": {}, "ios": {}, "js": {}, "linux": {}, "netbsd": {},
	"openbsd": {}, "plan9": {}, "solaris": {}, "wasip1": {}, "windows": {},
}

var architectures = map[string]struct{}{
	"386": {}, "amd64": {}, "arm": {}, "arm64": {}, "loong64": {},
	"mips": {}, "mips64": {}, "mips64le": {}, "mipsle": {}, "ppc64": {},
	"ppc64le": {}, "riscv64": {}, "s390x": {}, "wasm": {},
}

func isOS(value string) bool {
	_, ok := operatingSystems[value]
	return ok
}

func isArch(value string) bool {
	_, ok := architectures[value]
	return ok
}

func lastToken(value string) string {
	parts := strings.Split(value, "_")
	return parts[len(parts)-1]
}

func isGeneratedPart(path string) bool {
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	stem, _ = peelSuffix(stem, "_test")
	tokens := strings.Split(stem, "_")
	if len(tokens) > 1 {
		last := tokens[len(tokens)-1]
		if _, ok := buildTargets[last]; ok {
			stem = strings.TrimSuffix(stem, "_"+last)
			if prior := lastToken(stem); isOS(prior) && isArch(last) {
				stem = strings.TrimSuffix(stem, "_"+prior)
			}
		}
	}
	parts := strings.Split(stem, "_part")
	return len(parts) == 2 && parts[0] != "" && allDigits(parts[1])
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
