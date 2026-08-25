package languageprofile

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

func ObserveRuntime(request Request) Receipt { return Observe(request, RuntimeMeasurer{}) }

func Observe(request Request, measurer Measurer) Receipt {
	receipt := Receipt{
		Schema: ReceiptSchema, Filename: request.Filename, Entry: request.Entry,
		SourceDigest: digestBytes([]byte(request.Source)), Samples: []Sample{}, Effects: Effects{},
		Runner: Runner{GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH},
		NotClaimed: DefaultNonClaims(),
	}
	if strings.TrimSpace(request.Filename) == "" || request.Source == "" || strings.TrimSpace(request.Entry) == "" ||
		request.Samples < 1 || request.Samples > MaximumSamples || measurer == nil {
		return closeReceipt(receipt, "EXACT", "PROFILE_REQUEST_INVALID")
	}
	for sequence := 1; sequence <= request.Samples; sequence++ {
		execution, measurement := measurer.Measure(func() sourceexecution.Receipt {
			return sourceexecution.Execute(sourceexecution.Request{
				Filename: request.Filename, Source: request.Source, Entry: request.Entry,
			})
		})
		if execution.Decision != "PASS" {
			receipt.SourceDigest, receipt.SemanticDigest = execution.SourceDigest, execution.SemanticDigest
			return closeReceipt(receipt, "EXACT", execution.Reason)
		}
		if sequence == 1 {
			receipt.SourceDigest, receipt.SemanticDigest = execution.SourceDigest, execution.SemanticDigest
			receipt.ProfiledEntry = execution.Entry
		} else if execution.SourceDigest != receipt.SourceDigest || execution.SemanticDigest != receipt.SemanticDigest ||
			!reflect.DeepEqual(execution.Entry, receipt.ProfiledEntry) {
			return closeReceipt(receipt, "EXACT", "PROFILE_SUBJECT_DRIFT")
		}
		receipt.Samples = append(receipt.Samples, Sample{
			Sequence: sequence, Decision: execution.Decision, ExecutionDigest: execution.Digest,
			WallNanoseconds: measurement.WallNanoseconds, TotalAllocBytes: measurement.TotalAllocBytes,
		})
	}
	receipt.Summary = summarize(request.Samples, receipt.Samples)
	if receipt.Summary.WallObservations != request.Samples || receipt.Summary.AllocationObservations != request.Samples {
		return closeReceipt(receipt, "LOWER_RESOLUTION", "PROFILE_MEASUREMENT_UNKNOWN")
	}
	if receipt.Summary.ExecutionDigestVariants != 1 {
		return closeReceipt(receipt, "EXACT", "PROFILE_EXECUTION_DRIFT")
	}
	receipt.Decision, receipt.Resolution, receipt.Reason = "PASS", RunnerScopedResolution, "LANGUAGE_PROFILE_OBSERVED"
	return seal(receipt)
}

func closeReceipt(receipt Receipt, resolution, reason string) Receipt {
	receipt.Decision, receipt.Resolution, receipt.Reason = "FAIL_CLOSED", resolution, reason
	return seal(receipt)
}
