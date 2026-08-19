package coupling

import (
	"strconv"
	"strings"
)

func configCanonical(config Config) string {
	var builder strings.Builder
	field(&builder, ConfigSchemaV1)
	field(&builder, config.RegistryDigest)
	field(&builder, config.ToolchainDigest)
	field(&builder, config.ProfileDigest)
	field(&builder, config.SnapshotDigest)
	field(&builder, config.ExpectedProviderDigest)
	field(&builder, config.ExpectedObserverDigest)
	field(&builder, BaselineSchemaV1)
	field(&builder, strconv.FormatBool(config.Baseline.FullSuiteRequired))
	field(&builder, config.Baseline.Digest)
	field(&builder, strconv.FormatBool(config.ExternalReceiptRequired))
	return builder.String()
}
func baselineCanonical(baseline BaselineConfig) string {
	var builder strings.Builder
	field(&builder, BaselineSchemaV1)
	field(&builder, strconv.FormatBool(baseline.FullSuiteRequired))
	return builder.String()
}
func applicabilityCanonical(proof ApplicabilityProof) string {
	var builder strings.Builder
	field(&builder, AuthorityContextSchemaV1)
	field(&builder, proof.Schema)
	field(&builder, proof.RegistryDigest)
	field(&builder, proof.ToolchainDigest)
	field(&builder, proof.ProfileDigest)
	field(&builder, proof.SnapshotDigest)
	field(&builder, strconv.FormatBool(proof.AllowsEmpty))
	return builder.String()
}
func authorityCanonical(authority AuthorityContext) string {
	var builder strings.Builder
	field(&builder, AuthorityContextSchemaV1)
	field(&builder, authority.Schema)
	field(&builder, registryCanonical(authority.Registry))
	field(&builder, authority.Registry.Digest)
	field(&builder, authority.ToolchainDigest)
	field(&builder, authority.ProfileDigest)
	field(&builder, authority.SnapshotDigest)
	field(&builder, authority.ExpectedProviderDigest)
	field(&builder, authority.ExpectedObserverDigest)
	field(&builder, baselineCanonical(authority.Baseline))
	field(&builder, strconv.FormatBool(authority.ExternalReceiptRequired))
	if authority.Applicability != nil {
		field(&builder, applicabilityCanonical(*authority.Applicability))
		field(&builder, authority.Applicability.Digest)
	} else {
		field(&builder, "")
	}
	return builder.String()
}
