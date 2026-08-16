package couplingmanifest

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func resultForResolution(err error, authority RegistrySourceMap) (Manifest, error) {
	typed, ok := err.(*Error)
	if !ok {
		return unknownResult(CodeUnknownChangedSurface, err.Error(), authority)
	}
	if typed.Status == StatusFailClosed {
		return failResult(typed, authority)
	}
	return unknownResult(typed.Code, typed.Message, authority)
}

func failError(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Status: StatusFailClosed, FullSuiteRequired: true}
}

func unknownError(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Status: StatusUnknown, FullSuiteRequired: true}
}

func unknownResult(code ErrorCode, message string, authority RegistrySourceMap) (Manifest, error) {
	manifest := incompleteManifest(StatusUnknown, code, authority)
	return sealResult(manifest, unknownError(code, "%s", message))
}

func failResult(err error, authority RegistrySourceMap) (Manifest, error) {
	typed, ok := err.(*Error)
	if !ok {
		typed = failError(CodeInvalidManifest, "%v", err)
	}
	manifest := incompleteManifest(StatusFailClosed, typed.Code, authority)
	return sealResult(manifest, typed)
}

func incompleteManifest(status Status, code ErrorCode, authority RegistrySourceMap) Manifest {
	manifest := Manifest{
		Schema: SchemaV1, Complete: false, ZeroChange: false,
		Entries: []ManifestEntry{}, Status: status, FullSuiteRequired: true,
		ReasonCode: code, ResolvedSurfaceIDs: []semantic.ID{}, Counts: ComponentCounts{}, Work: Work{}, statsKnown: true,
	}
	setAuthorityDigests(&manifest, authority)
	return manifest
}

func setAuthorityDigests(manifest *Manifest, authority RegistrySourceMap) {
	if digest, err := rawDigest(authority.RegistryDigest); err == nil {
		manifest.RegistryDigest = digest
	}
	if digest, err := rawDigest(authority.SourceMapDigest); err == nil {
		manifest.SourceMapDigest = digest
	}
	if digest, err := rawDigest(authority.ToolchainDigest); err == nil {
		manifest.ToolchainDigest = digest
	}
	if digest, err := rawDigest(authority.ProfileDigest); err == nil {
		manifest.ProfileDigest = digest
	}
}
