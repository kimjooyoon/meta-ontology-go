package valuecatalog

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

type observation struct {
	baselineSource, candidateSource            []byte
	actualCore, baselineCore                   coreObservation
	baseline, extension                        ProgramResult
	beforeReason, unknownReason                string
	extensionPresent, coreFingerprintSensitive bool
}

func observe(path string, source []byte) (observation, error) {
	baselineSource, candidateSource, err := catalogSources(source)
	if err != nil {
		return observation{}, err
	}
	actualCore, err := observeCore(path, source)
	if err != nil {
		return observation{}, err
	}
	baselineCore, err := observeCore(path, baselineSource)
	if err != nil {
		return observation{}, err
	}
	baselineProgram, err := valueexecution.Compile(path, source, BaselineActivity)
	if err != nil {
		return observation{}, fmt.Errorf("baseline compile: %w", err)
	}
	_, beforeErr := valueexecution.Compile(path, baselineSource, ExtensionActivity)
	beforeReason := valueexecution.Reason(beforeErr)
	if beforeReason != valueexecution.ReasonProgramMissing {
		return observation{}, fmt.Errorf("extension baseline reason = %s", beforeReason)
	}
	extensionProgram, extensionErr := valueexecution.Compile(path, source, ExtensionActivity)
	extensionReason := valueexecution.Reason(extensionErr)
	if extensionErr != nil && extensionReason != valueexecution.ReasonProgramMissing {
		return observation{}, fmt.Errorf("extension compile reason = %s", extensionReason)
	}
	_, unknownErr := valueexecution.Compile(path, unknownExtensionSource(candidateSource), ExtensionActivity)
	observed := observation{
		baselineSource: baselineSource, candidateSource: candidateSource,
		actualCore: actualCore, baselineCore: baselineCore,
		baseline: executeProgram(baselineProgram, 1), beforeReason: beforeReason,
		unknownReason: valueexecution.Reason(unknownErr), extensionPresent: extensionErr == nil,
		coreFingerprintSensitive: actualCore.fingerprint != baselineCore.fingerprint,
	}
	if extensionErr == nil {
		observed.extension = executeProgram(extensionProgram, 2)
	} else {
		observed.extension = ProgramResult{Activity: ExtensionActivity, CompileReason: extensionReason}
	}
	return observed, nil
}
