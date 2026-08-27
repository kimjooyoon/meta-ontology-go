package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func sourceSemanticDigest(cases []declaredCase) string {
	var canonical strings.Builder
	for _, item := range cases {
		fmt.Fprintf(&canonical, "%s|%d|%d|%s|%d|%t|%s|%s\n", item.ID,
			item.Observation.Required, item.Observation.Observed, item.Observation.Reason,
			item.Observation.RepositoryWrites, item.Observation.MutationAuthority,
			item.ClaimID, item.ExpectedClaimState)
	}
	return digestText(canonical.String())
}

func digestText(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}
