package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func checkPullRequestRoute(head, base string) error {
	if !verify.IsFoundationPromotionRoute(head, base) {
		return verify.CheckPullRequestPolicy(head, base)
	}
	prNumber, err := strconv.ParseInt(os.Getenv("GOOO_PR_NUMBER"), 10, 64)
	if err != nil {
		return fmt.Errorf("foundation promotion PR number is malformed")
	}
	results := make(map[string]string)
	for _, pair := range strings.Split(os.Getenv("GOOO_NON_ROUTE_RESULTS"), ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			results[parts[0]] = parts[1]
		}
	}
	return verify.CheckFoundationPromotionPolicy(verify.FoundationPromotionPolicyInput{
		Repository:     os.Getenv("GOOO_REPOSITORY"),
		PRNumber:       prNumber,
		HeadBranch:     head,
		BaseBranch:     base,
		BaseSHA:        os.Getenv("GOOO_PR_BASE_SHA"),
		HeadSHA:        os.Getenv("GOOO_EXPECTED_HEAD"),
		HumanDecision:  os.Getenv("GOOO_HUMAN_DECISION"),
		NonRouteResult: results,
	})
}
