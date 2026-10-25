package main

import (
	"context"
	"encoding/json"
	"io"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

type packageWallDiagnostic struct {
	Schema         string              `json:"schema"`
	Operation      string              `json:"operation"`
	Decision       string              `json:"decision"`
	ManifestDigest string              `json:"manifest_digest,omitempty"`
	InputsDigest   string              `json:"inputs_digest,omitempty"`
	CasesDigest    string              `json:"cases_digest,omitempty"`
	SourceSHA256   string              `json:"execution_source_sha256,omitempty"`
	Wall           bodytiming.Snapshot `json:"wall"`
	Scope          string              `json:"scope"`
}

// Timing is an optional view. No recorder or clock is allocated when disabled.
func packageTimingContext(ctx context.Context, args []string) (context.Context, []string, *bodytiming.Recorder) {
	if !slices.Contains(args, "--timing") {
		return ctx, args, nil
	}
	clean := make([]string, 0, len(args))
	for _, arg := range args {
		if arg != "--timing" {
			clean = append(clean, arg)
		}
	}
	recorder := bodytiming.NewRecorder()
	return bodytiming.WithRecorder(ctx, recorder), clean, recorder
}

func writePackageTiming(stderr io.Writer, recorder *bodytiming.Recorder, operation string, receipt packageExecutionReceipt) {
	if recorder == nil {
		return
	}
	diagnostic := packageWallDiagnostic{Schema: "gooo/package-wall-diagnostic/v1", Operation: operation,
		Decision: receipt.Decision, ManifestDigest: receipt.ManifestDigest, InputsDigest: receipt.InputsDigest,
		CasesDigest: receipt.CasesDigest, Wall: recorder.Snapshot(),
		Scope: "Optional current-command monotonic wall intervals through stdout publication, before diagnostic output; " +
			"unassigned time includes workspace construction, origin binding, receipt encoding and bookkeeping. " +
			"Phases may repeat across sequential compositions. Missing phases are unobserved. " +
			"Process sub-timings overlap these intervals; do not add them. This view grants no execution authority."}
	if receipt.Result != nil {
		diagnostic.SourceSHA256 = receipt.Result.SourceSHA256
	}
	// A failed diagnostic sink must not change execution or receipt publication.
	_ = json.NewEncoder(stderr).Encode(diagnostic)
}
