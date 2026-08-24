package sourceauthorityupstream

const (
	gomacroRepository = "cosmos72/gomacro"
	gomacroRevision   = "cf0d4bf32da393dbda97e3572f216731013ffa55"
	gomacroPath       = "README.md"
	gomacroURL        = "https://raw.githubusercontent.com/cosmos72/gomacro/cf0d4bf32da393dbda97e3572f216731013ffa55/README.md"
	gomacroDigest     = "sha256:29362aa311de0f24c66f41cc65a8b6ffd996baf37e048b5a72db63172aae5bf2"
)

func GomacroPolicy() Policy {
	return Policy{
		SourceRef:      "gomacro-readme-title",
		AuthorityRef:   "gomacro-readme-title-authority",
		URL:            gomacroURL,
		Authority:      Authority{Repository: gomacroRepository, Revision: gomacroRevision, Path: gomacroPath},
		Selection:      Selection{StartLine: 1, EndLine: 1},
		ExpectedDigest: gomacroDigest,
		ExpectedBytes:  77,
	}
}

func GomacroRequest(subjectSHA string) Request {
	policy := GomacroPolicy()
	return Request{
		Schema:       RequestSchema,
		SubjectSHA:   subjectSHA,
		SourceRef:    policy.SourceRef,
		AuthorityRef: policy.AuthorityRef,
		URL:          policy.URL,
		Authority:    policy.Authority,
		Selection:    policy.Selection,
	}
}
