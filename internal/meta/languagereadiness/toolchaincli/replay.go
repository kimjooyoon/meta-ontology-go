package toolchaincli

import cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"

func deterministicReplayEqual(first, replay cliruntime.Observation) bool {
	return cliruntime.DeterministicReplayEqual(first, replay)
}
