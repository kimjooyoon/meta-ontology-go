package packageexecution

func sourceIdentityDigest(sources []Source) string {
	return digestValue(sources)
}
