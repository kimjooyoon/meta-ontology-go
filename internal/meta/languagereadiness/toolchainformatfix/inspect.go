package toolchainformatfix

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

type formatJSON struct {
	Schema, Command, Status, File, Source, SourceDigest, FormattedDigest string
	Changed, DirectWrites                                               bool
}

func inspectOutput(operation string, observed cliruntime.Observation) (bool, int, int) {
	switch operation {
	case "FORMAT_TEXT":
		return observed.Stdout == canonical && observed.Stderr == "", 0, 0
	case "FORMAT_JSON":
		value := struct {
			Schema, Command, Status, File, Source string
			Changed                               bool
			DirectWrites                          int
		}{}
		if json.Unmarshal([]byte(observed.Stdout), &value) != nil {
			return false, 0, 0
		}
		ok := value.Schema == "gooo/format-report/v1" && value.Command == "format" &&
			value.Status == "formatted" && value.File == unformattedPath &&
			value.Source == canonical && value.Changed && value.DirectWrites == 0 &&
			observed.Stderr == ""
		return ok, 1, 0
	case "FORMAT_CHECK":
		return observed.Stdout == "canonical: "+canonicalPath+"\n" && observed.Stderr == "", 0, 0
	case "FIX_JSON_CHANGED":
		return inspectPlan(observed, unformattedPath, unformatted, formatfix.DecisionChangePlanned)
	case "FIX_JSON_FIXED":
		return inspectPlan(observed, canonicalPath, canonical, formatfix.DecisionFixedPoint)
	case "FIX_TEXT":
		plan := formatfix.Build(unformattedPath, unformatted)
		expected := fmt.Sprintf("%s: %s edits=1 writes=0 digest=%s\n",
			plan.Decision, unformattedPath, plan.PlanDigest)
		return observed.Stdout == expected && observed.Stderr == "", 0, 0
	default:
		return inspectGuardrail(operation, observed), 0, 0
	}
}

func inspectPlan(observed cliruntime.Observation, filename, source string,
	decision formatfix.Decision) (bool, int, int) {
	plan := formatfix.Plan{}
	if json.Unmarshal([]byte(observed.Stdout), &plan) != nil || formatfix.Validate(plan) != nil {
		return false, 0, 0
	}
	expected := formatfix.Build(filename, source)
	ok := plan.PlanDigest == expected.PlanDigest && plan.Decision == decision &&
		plan.DirectWrites == 0 && observed.Stderr == ""
	return ok, 1, 1
}

func inspectGuardrail(operation string, observed cliruntime.Observation) bool {
	expected := map[string]string{
		"FORMAT_REQUIRED":  "usage: gooo format [--check] [--json] <file.gooo>\n",
		"FORMAT_MALFORMED": "gooo: " + malformedPath + ": FORMAT_FIX_SOURCE_UNKNOWN\n",
		"FIX_MALFORMED":    "gooo: " + malformedPath + ": FORMAT_FIX_SOURCE_UNKNOWN\n",
		"FORMAT_USAGE":     "usage: gooo format [--check] [--json] <file.gooo>\n",
		"FIX_USAGE":        "usage: gooo fix [--json] <file.gooo>\n",
		"FIX_FLAG":         "usage: gooo fix [--json] <file.gooo>\n",
	}
	return observed.Stdout == "" && observed.Stderr == expected[operation]
}
