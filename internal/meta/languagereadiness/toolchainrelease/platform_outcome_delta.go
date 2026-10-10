package toolchainrelease

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const outcomeDeltaObservations = "docs/research/workflow-outcome-delta-20261011"

// These are saved observations. Running the comparator does not repeat the
// original construction, model prediction, or generated-program execution.
func smokeOutcomeDelta(binary, work string, input BuildInput) error {
	return runOutcomeDeltaSmoke(binary, work, input, commandOutput)
}

func runOutcomeDeltaSmoke(binary, work string, input BuildInput,
	run func(string, []string, string, ...string) ([]byte, error)) error {
	root := filepath.Join(input.Root, outcomeDeltaObservations)
	for _, name := range []string{"before", "after", "replay"} {
		if err := unpackOutcomeInput(root, work, name); err != nil {
			return err
		}
	}
	for _, sample := range []struct{ after, expected string }{
		{"after", "change"}, {"replay", "saved-replay-delta"},
	} {
		args := []string{"body-outcomes-delta", "--before", filepath.Join(work, "before.json"),
			"--after", filepath.Join(work, sample.after+".json"), "--json"}
		raw, runErr := run(input.Root, nil, binary, args...)
		raw, err := retainLanguageCommandOutput(input, "outcome-delta-"+sample.expected, raw, runErr)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_OUTCOME_DELTA %s: %w", sample.expected, err)
		}
		expected, err := os.ReadFile(filepath.Join(root, sample.expected+".json"))
		if err != nil {
			return err
		}
		if !bytes.Equal(raw, expected) {
			return fmt.Errorf("TOOLCHAIN_RELEASE_OUTCOME_DELTA %s: saved comparison differs", sample.expected)
		}
	}
	return nil
}

func unpackOutcomeInput(root, work, name string) error {
	file, err := os.Open(filepath.Join(root, name+".json.gz"))
	if err != nil {
		return err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()
	const limit = 32 << 20
	raw, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return err
	}
	if len(raw) > limit {
		return fmt.Errorf("saved outcome input exceeds %d bytes", limit)
	}
	return os.WriteFile(filepath.Join(work, name+".json"), raw, 0o600)
}
