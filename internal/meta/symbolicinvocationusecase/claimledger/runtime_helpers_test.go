package claimledger

import "encoding/json"

func runtimeEvidence(subject, observationDigest string, peakRSS int) []byte {
	evidence := map[string]any{
		"schema":      "gooo/runtime-measurement-evidence/v1",
		"subject_sha": subject,
		"coordinate": map[string]any{
			"stage": "OBSERVE",
			"step":  "capture-peak-rss",
		},
		"source": map[string]any{"observation_digest": observationDigest},
		"producer": map[string]any{
			"tool": "GNU time", "binary_path": "/usr/bin/time",
			"binary_digest": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			"version_digest": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		},
		"measurement": map[string]any{
			"target": "symbolic-reader-request-observer", "peak_rss_kib": peakRSS, "unit": "KiB",
		},
		"effects": map[string]any{"repository_writes": 0, "mutation_authority": false},
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		panic(err)
	}
	return encoded
}
