package toolchainconformance

func inspectIdentity(definition SurfaceDefinition, envelope artifactEnvelope,
	expectedHead string, summary *Summary) string {
	if envelope.Schema != definition.Schema {
		summary.SchemaMismatches++
	}
	if envelope.Decision != DecisionPass {
		summary.DecisionMismatches++
	}
	if envelope.Resolution != ResolutionExact {
		summary.ResolutionDescents++
	}
	head := envelope.HeadSHA
	if head == "" {
		head = envelope.Source.ExpectedHeadSHA
	} else if envelope.Source.ExpectedHeadSHA != "" &&
		envelope.Source.ExpectedHeadSHA != head {
		summary.HeadMismatches++
	}
	if head == expectedHead && validHead(head) {
		summary.HeadBindings++
	} else {
		summary.HeadMismatches++
	}
	if !validDigest(envelope.ReportDigest) {
		summary.DigestFailures++
	}
	if envelope.RepositoryWrites != 0 {
		summary.RepositoryWrites += envelope.RepositoryWrites
	}
	if envelope.MutationAuthorized {
		summary.MutationAuthorities++
	}
	return head
}
