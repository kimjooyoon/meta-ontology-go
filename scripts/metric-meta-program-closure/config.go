package main

import "flag"

type config struct {
	repository, subjectSHA                   string
	programPath, sourcePath, verificationPath string
	artifactName, artifactDigest, artifactURL string
	outputDir                                string
	runID, artifactID                        int64
	runAttempt                               int
}

func parseConfig() config {
	var value config
	flag.StringVar(&value.repository, "repository", "", "owner/repository")
	flag.StringVar(&value.subjectSHA, "subject-sha", "", "exact subject commit")
	flag.Int64Var(&value.runID, "run-id", 0, "GitHub Actions run id")
	flag.IntVar(&value.runAttempt, "run-attempt", 0, "GitHub Actions run attempt")
	flag.Int64Var(&value.artifactID, "artifact-id", 0, "program artifact id")
	flag.StringVar(&value.artifactName, "artifact-name", "", "program artifact name")
	flag.StringVar(&value.artifactDigest, "artifact-digest", "", "program artifact sha256")
	flag.StringVar(&value.artifactURL, "artifact-url", "", "program artifact URL")
	flag.StringVar(&value.programPath, "program", "", "program JSON")
	flag.StringVar(&value.sourcePath, "source", "", "program Gooo source")
	flag.StringVar(&value.verificationPath, "verification", "", "program verification JSON")
	flag.StringVar(&value.outputDir, "out", "", "explicit artifact output directory")
	flag.Parse()
	return value
}
