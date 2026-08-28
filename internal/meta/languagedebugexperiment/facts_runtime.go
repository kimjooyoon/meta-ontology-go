package languagedebugexperiment

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func validMeasurement(value Measurement, expectedName string) bool {
	return value.Name == expectedName && value.Executed && value.WallNS > 0 && value.WallMS == (value.WallNS+999999)/1000000 && value.WallMS > 0 && value.PeakRSSKiB > 0 && value.CacheState != ""
}

func runtimeUncertainty(input Input) *Uncertainty {
	if len(input.RuntimeObservations) != input.Contract.ExpectedResourceObservations { value := unknownCase("RESOURCE_OBSERVED", "READ_RUNTIME_OBSERVATION", "RUNTIME_OBSERVATION_MISSING", "MISSING_EVIDENCE", "REEXECUTE_DEBUG_PATH_WITH_RESOURCES", "RUNTIME_OBSERVATIONS"); return &value }
	if !validMeasurement(input.Build, "debug-producer-build") || !validMeasurement(input.EvaluatorBuild, "debug-evaluator-build") || !validMeasurement(input.Test, "debug-relevant-tests") { value := unknownCase("RESOURCE_OBSERVED", "READ_BUILD_TEST_MEASUREMENTS", "BUILD_OR_TEST_RESOURCE_MISSING", "MISSING_EVIDENCE", "RECORD_CI_MEASUREMENTS", "BUILD_TEST_RECEIPTS"); return &value }
	positive := []languagedebug.Receipt{input.First, input.Second}
	for index, runtime := range input.RuntimeObservations {
		if runtime.Run != index+1 || runtime.RuntimeReceiptSchema != RuntimeReceiptSchema || runtime.Runner == "" || !strings.Contains(runtime.Toolchain, "go1.27") || !validDigest(runtime.SourceRawDigest) || !validDigest(runtime.SourceSemanticDigest) || !validDigest(runtime.BinaryDigest) || len(runtime.Arguments) == 0 || !validSHA(runtime.SubjectSHA) || runtime.SubjectSHA != input.SubjectSHA || !validDigest(runtime.OutputDigest) || runtime.WallNS <= 0 || runtime.WallMS <= 0 || runtime.WallMS != (runtime.WallNS+999999)/1000000 || runtime.PeakRSSKiB <= 0 || runtime.BinaryDigest != input.ExecutableDigest || runtime.SourceRawDigest != positive[index].SourceDigest || runtime.SourceSemanticDigest != positive[index].SemanticDigest {
			value := unknownCase("RESOURCE_OBSERVED", "VALIDATE_RUNTIME_OBSERVATION", "RUNTIME_FIELD_MISSING_OR_CONTRADICTED", "INCOMPLETE_EVIDENCE", "REEXECUTE_DEBUG_PATH_WITH_COMPLETE_RECEIPT", "RUNNER_RESOURCE_RECEIPT"); return &value
		}
	}
	return nil
}
