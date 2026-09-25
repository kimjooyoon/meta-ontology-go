package verify

import (
	"fmt"
	"strings"
)

const (
	FoundationPromotionRepository    = "kimjooyoon/meta-ontology-go"
	FoundationPromotionPRNumber      = int64(602)
	FoundationPromotionHeadBranch    = "agent/foundation-discovery-recovery-20260830"
	FoundationPromotionBaseBranch    = "main"
	FoundationPromotionBaseSHA       = "cd9727af80f5118405290d3be96890c18e1529c0"
	FoundationPromotionHumanDecision = "ALLOW_ONE_TIME_FOUNDATION_PROMOTION_FOR_PR_602_WITH_NORMAL_ROUTES_UNCHANGED"
)

type FoundationPromotionPolicyInput struct {
	Repository     string
	PRNumber       int64
	HeadBranch     string
	BaseBranch     string
	BaseSHA        string
	HeadSHA        string
	HumanDecision  string
	NonRouteResult map[string]string
}

func IsFoundationPromotionRoute(head, base string) bool {
	return head == FoundationPromotionHeadBranch && base == FoundationPromotionBaseBranch
}

func CheckFoundationPromotionPolicy(input FoundationPromotionPolicyInput) error {
	if input.Repository != FoundationPromotionRepository || input.PRNumber != FoundationPromotionPRNumber ||
		input.HeadBranch != FoundationPromotionHeadBranch || input.BaseBranch != FoundationPromotionBaseBranch ||
		input.BaseSHA != FoundationPromotionBaseSHA || !validFoundationSHA(input.HeadSHA) {
		return fmt.Errorf("foundation promotion identity is not the exact authorized PR #602 tuple")
	}
	if input.HumanDecision != FoundationPromotionHumanDecision {
		return fmt.Errorf("foundation promotion lacks the exact human decision record")
	}
	for _, job := range []string{"format", "vet", "test", "race", "semantic"} {
		if input.NonRouteResult[job] != "success" {
			return fmt.Errorf("foundation promotion requires non-route job %q to pass", job)
		}
	}
	return nil
}

func validFoundationSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
