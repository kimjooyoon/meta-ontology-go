package verify

import (
	"os"
	"strings"
	"testing"
)

// Keep the actual orchestration and compile wrapper. Only the projection
// callbacks are controlled; the independent Go assertion executes normally.
func finalProjectionHarness(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../../scripts/semantic-conformance.sh")
	if err != nil {
		t.Fatal(err)
	}
	_, rest, found := strings.Cut(string(raw), "run_projected_compile() {\n")
	body, _, ended := strings.Cut(rest, "\n}\n")
	start := "\nset +e\nrun_extraction_pass \"$projection_work/extraction-report.json\" false "
	_, rest, started := strings.Cut(string(raw), start)
	pipeline, _, closed := strings.Cut(rest, "\n# Stage 0 keeps")
	if !found || !ended || !started || !closed {
		t.Fatal("projected validation orchestration not found")
	}
	return finalProjectionAdapters + "run_projected_compile() {\n" + body + "\n}\n" + start + pipeline + "\n"
}

const finalProjectionAdapters = `set -euo pipefail
repo_root="$1"
cd "$repo_root"
projection_work="$repo_root/projection"
storage_root="$repo_root"
head_sha=fixture
mkdir -p "$projection_work"
printf '{}\n' > "$projection_work/projected-compile.json"
go() {
  if [[ "$1" == test ]]; then
    printf 'compile\n' >> "$repo_root/events"
    "$REAL_GO" "$@"
    return $?
  fi
  case "$2" in
    ./scripts/line-metrics)
      blockers=1
      if [[ "$CLOSURE" == closed && -f "$repo_root/rewritten" ]]; then blockers=0; fi
      jq -n --argjson value "$blockers" '{meta:{indicators:[{metric_id:"gooo.metric.source.function-lines.v1",applicability:"APPLICABLE",role:"DRIVER",value:$value,limit:0}]}}'
      ;;
    ./bootstrap/meta-binding-witness|./bootstrap/logical-split-planner|./bootstrap/line-density-rewriter|./scripts/source-splitter)
      ;;
    *) printf 'unhandled projection command: %s\n' "$2" >&2; return 91 ;;
  esac
}
run_extraction_pass() {
  if [[ "$2" == true ]]; then
    printf 'package sample\nfunc Value() int { return %s }\n' "$VALUE" > "$repo_root/sample.go"
    printf 'rewrite\n' >> "$repo_root/events"
    touch "$repo_root/rewritten"
  fi
  printf '{}\n' > "$1"
  EXTRACTION_PARTIAL=0
}
write_extraction_result() {
  jq -n --slurpfile compiled "$projection_work/projected-compile.json" '{projected_compile:$compiled[0]}' > "$2"
}
`
