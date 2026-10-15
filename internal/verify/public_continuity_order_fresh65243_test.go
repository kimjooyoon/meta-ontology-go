package verify

import (
	"os"
	"strings"
	"testing"
)

func TestPublicContinuityCompletesProducersBeforeTransport65243(t *testing.T) {
	path := os.Getenv("GOOO_CONTINUITY_ORDER_WORKFLOW65243")
	if path == "" {
		path = "../../.github/workflows/self-improvement-public-continuity.yml"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !publicContinuityProducerOrder65243(string(raw)) {
		t.Fatal("artifact transport precedes a declared source producer")
	}
	for _, marker := range []string{
		"uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a",
		"        if: always()",
		"          if-no-files-found: error",
		"          retention-days: 30",
	} {
		if strings.Count(string(raw), marker) != 5 {
			t.Fatalf("five strict original upload declarations required: %s", marker)
		}
	}
	if strings.Contains(string(raw), "continue-on-error:") {
		t.Fatal("transport failure was masked")
	}
}

func TestPublicContinuityOrderRejectsSkippedOrLateProducers65243(t *testing.T) {
	stages := publicContinuityStages65243()
	var all strings.Builder
	for _, name := range stages {
		all.WriteString("      - name: " + name + "\n")
	}
	complete := all.String() + "      - name: Upload complete dossier\n"
	if !publicContinuityProducerOrder65243(complete) {
		t.Fatal("all source stages before transport were rejected")
	}
	for _, source := range []string{
		"      - name: Upload too early\n" + all.String(),
		strings.Replace(complete, "      - name: "+stages[7]+"\n", "", 1),
		complete + "      - name: " + stages[0] + "\n",
	} {
		if publicContinuityProducerOrder65243(source) {
			t.Fatal("missing or late producer was accepted")
		}
	}
}

func publicContinuityProducerOrder65243(source string) bool {
	required := publicContinuityStages65243()
	seen := make(map[string]int, len(required))
	transport := false
	for line := range strings.SplitSeq(source, "\n") {
		name, step := strings.CutPrefix(line, "      - name: ")
		if !step {
			continue
		}
		if strings.HasPrefix(name, "Upload ") {
			transport = true
			continue
		}
		for _, stage := range required {
			if name == stage {
				if transport {
					return false
				}
				seen[stage]++
			}
		}
	}
	for _, stage := range required {
		if seen[stage] != 1 {
			return false
		}
	}
	return transport
}

func publicContinuityStages65243() []string {
	return []string{
		"Run the public candidate to decision to certificate to consumption path",
		"Verify the six lowered-policy continuity cases and exact artifacts",
		"Accept and measure the generated project utility",
		"Reuse an authorized generated-test receipt across public commands",
		"Run the meta-defined public orchestration protocol",
		"Run graph-driven partial test reuse from the v14 boundary",
		"Reject labels-only and contradictory orchestration admission",
		"Run counterexample-driven semantic resolution repair",
	}
}
