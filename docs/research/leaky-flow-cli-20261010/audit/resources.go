package main

import (
	"path/filepath"
	"regexp"
	"strconv"
)

func processResources(root string) map[string]any {
	times := regexp.MustCompile(`([0-9.]+) real\s+([0-9.]+) user\s+([0-9.]+) sys`)
	memory := regexp.MustCompile(`([0-9]+)\s+maximum resident set size`)
	var seconds [3]float64
	var peak int64
	count := 0
	for _, f := range forms {
		for _, command := range []string{"preflight", "deterministic", "initial", "feedback", "construction", "replay"} {
			text := string(read(filepath.Join(root, f.Name), command+".time-stderr"))
			t, m := times.FindStringSubmatch(text), memory.FindStringSubmatch(text)
			must(len(t) == 4 && len(m) == 2, "original process measurements")
			for i := range 3 {
				v, err := strconv.ParseFloat(t[i+1], 64)
				must(err == nil && v >= 0, "nonnegative process time")
				seconds[i] += v
			}
			bytes, err := strconv.ParseInt(m[1], 10, 64)
			must(err == nil && bytes > 0, "RSS bytes")
			peak = max(peak, bytes)
			count++
		}
	}
	return map[string]any{"commands": count, "sum_wall_seconds": seconds[0], "sum_user_seconds": seconds[1],
		"sum_system_seconds": seconds[2], "max_rss_bytes": peak,
		"scope": "sum of rounded per-command process observations including native tooling; host utilization delta not measured"}
}
