package proofchoicealgebra

func Evaluate(path string, source []byte) Receipt {
	bundle, issues := parseBundle(path, source)
	receipt := Receipt{
		Schema: Schema, Decision: Pass, Reason: "PROOF_CHOICES_COMPOSED", Resolution: Exact,
		SourcePath: path, SourceDigest: digestSource(source), FixedDenom: FixedDenominator,
		Items: bundle.Items, Transitions: bundle.Transitions,
		Effects: Effects{RepositoryWrites: 0, MutationAuthority: false},
	}
	receipt.Summary = summarize(bundle)
	receipt.Indicators = indicators(bundle)
	failure := firstIssue(issues)
	if failure == "" {
		failure = validateBundle(bundle)
	}
	if failure != "" {
		receipt.Decision, receipt.Reason, receipt.Resolution = FailClosed, failure, FailClosed
		for index := range receipt.Indicators {
			receipt.Indicators[index].Decision = FailClosed
		}
	}
	receipt.Summary.Unknowns = countUnknowns(bundle)
	receipt.Summary.Contradictions = countContradictions(bundle)
	receipt.Summary.Compositions = len(bundle.Items) + len(bundle.Transitions)
	digest, err := digestReceipt(receipt)
	if err != nil {
		receipt.Decision, receipt.Reason, receipt.Resolution = FailClosed, "RECEIPT_DIGEST_UNKNOWN", FailClosed
	} else {
		receipt.Digest = digest
	}
	return receipt
}

func firstIssue(issues []issue) string {
	if len(issues) == 0 {
		return ""
	}
	return issues[0].Reason
}
