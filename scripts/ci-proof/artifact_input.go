package main

type artifactInput struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size_bytes"`
	Expired    bool   `json:"expired"`
	Digest     string `json:"digest"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}
