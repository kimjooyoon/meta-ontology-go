package bodycodegen

import (
	"context"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

type tinyGoBodyFillResolver interface {
	Resolve(context.Context, decisionroute.Request) (decisionroute.Receipt, error)
}

// IRBodyFillOptions selects an optional local model for one body-fill call.
// A nil TinyGoProvider preserves the existing Laya/default behavior.
type IRBodyFillOptions struct {
	TinyGoProvider  *decisionroute.TinyGoProvider
	TinyModelLoadMS *float64
}

func tinyGoBodyFillOptions(candidates []IRBodyFillCandidate) ([]decisionroute.Option, error) {
	options := make([]decisionroute.Option, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		operation, err := tinyGoRootOperation(candidate.Expression)
		if err != nil {
			return nil, fmt.Errorf("candidate %q cannot map to a tiny_go operation: %w", candidate.ID, err)
		}
		if _, duplicate := seen[operation]; duplicate {
			return nil, fmt.Errorf("tiny_go candidate operations must be unique: %q", operation)
		}
		seen[operation] = struct{}{}
		options = append(options, decisionroute.Option{
			ID: candidate.ID, Operation: operation,
			Description: fmt.Sprintf("Emit exactly this expression: %s.", candidate.Expression),
		})
	}
	return options, nil
}

func tinyGoRootOperation(expression string) (string, error) {
	parsed, err := parser.ParseExpr(expression)
	if err != nil {
		return "", err
	}
	for {
		parenthesized, ok := parsed.(*ast.ParenExpr)
		if !ok {
			break
		}
		parsed = parenthesized.X
	}
	binary, ok := parsed.(*ast.BinaryExpr)
	if !ok {
		return "", fmt.Errorf("root expression must be one supported binary operation")
	}
	switch binary.Op {
	case token.ADD:
		return decisionroute.TinyGoOperationAdd, nil
	case token.SUB:
		return decisionroute.TinyGoOperationSubtract, nil
	case token.MUL:
		return decisionroute.TinyGoOperationMultiply, nil
	case token.LSS:
		return decisionroute.TinyGoOperationLessThan, nil
	case token.LEQ:
		return decisionroute.TinyGoOperationLessEqual, nil
	case token.EQL:
		return decisionroute.TinyGoOperationEqual, nil
	case token.LAND:
		return decisionroute.TinyGoOperationAnd, nil
	case token.LOR:
		return decisionroute.TinyGoOperationOr, nil
	default:
		return "", fmt.Errorf("root operator %q is outside the closed tiny_go operation set", binary.Op)
	}
}

func bodyFillProviderAccounting(receipt decisionroute.Receipt) (localPredictions, externalCalls int, externalCallsKnown bool) {
	switch receipt.Provider {
	case decisionroute.ProviderTinyGo:
		return 1, 0, true
	case "laya":
		return 0, 1, true
	case "deterministic":
		if receipt.FallbackReason == "NOT_CONFIGURED" {
			return 0, 0, true
		}
		return 0, 0, false
	default:
		return 0, 0, false
	}
}

func validTinyGoDecisionReceipt(receipt decisionroute.Receipt, expectedRequestSHA256 string) bool {
	switch receipt.TinyGoVariant {
	case "fp32", "ptq_ternary", "qat_ternary":
	default:
		return false
	}
	if len(receipt.TinyGoWeightsSHA256) != 64 || receipt.TinyGoWeightsSHA256 != strings.ToLower(receipt.TinyGoWeightsSHA256) {
		return false
	}
	if _, err := hex.DecodeString(receipt.TinyGoWeightsSHA256); err != nil {
		return false
	}
	if len(receipt.TinyGoMetadataSHA256) != 64 || receipt.TinyGoMetadataSHA256 != strings.ToLower(receipt.TinyGoMetadataSHA256) {
		return false
	}
	if _, err := hex.DecodeString(receipt.TinyGoMetadataSHA256); err != nil {
		return false
	}
	return receipt.RequestSHA256 == expectedRequestSHA256
}
