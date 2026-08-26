package feedbackpredecessor

func applyUseCase(input *Input, state string) {
	switch state {
	case "canonical":
	case "missing":
		input.Candidates = nil
	case "unsuccessful":
		input.Candidates[0].Conclusion = "failure"
		input.Candidates[0].ReceiptDigest = ""
	case "expired":
		input.Candidates[0].Expired = true
	case "malformed_receipt":
		input.Candidates[0].ReceiptDigest = "sha256:unknown"
	case "duplicate":
		input.Candidates = append(input.Candidates, input.Candidates[0])
	case "write_effect":
		input.Candidates[0].RepositoryWrites = 1
	case "noncanonical":
		input.Candidates[0].HeadBranch = "feature"
	}
}
