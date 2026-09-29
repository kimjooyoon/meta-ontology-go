package decisionroute

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const RequestSchema = "gooo/typed-decision-request/v1"
const ReceiptSchema = "gooo/typed-decision-receipt/v1"

type Request struct {
	Schema   string   `json:"schema"`
	State    string   `json:"state"`
	Question Question `json:"question"`
	Fallback string   `json:"fallback"`
}

type Question struct {
	ID           string   `json:"id"`
	Instructions string   `json:"instructions"`
	Options      []Option `json:"options"`
}

type Option struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type Receipt struct {
	Schema           string             `json:"schema"`
	Mode             string             `json:"mode"`
	Selected         string             `json:"selected"`
	FallbackReason   string             `json:"fallback_reason,omitempty"`
	Provider         string             `json:"provider"`
	Model            string             `json:"model,omitempty"`
	ModelRevision    string             `json:"model_revision,omitempty"`
	Routing          map[string]any     `json:"routing,omitempty"`
	Probabilities    map[string]float64 `json:"probabilities,omitempty"`
	Confidence       *float64           `json:"confidence,omitempty"`
	AnswerConfidence *float64           `json:"answer_confidence,omitempty"`
	RequestSHA256    string             `json:"request_sha256"`
}

func Validate(request Request) (string, error) {
	if request.Schema != RequestSchema || strings.TrimSpace(request.State) == "" ||
		strings.TrimSpace(request.Question.ID) == "" || strings.TrimSpace(request.Question.Instructions) == "" {
		return "", errors.New("schema, state, question id, and question instructions are required")
	}
	if len(request.Question.Options) < 2 {
		return "", errors.New("at least two decision options are required")
	}
	seen := make(map[string]bool, len(request.Question.Options))
	fallbackFound := false
	for _, option := range request.Question.Options {
		if strings.TrimSpace(option.ID) == "" || strings.TrimSpace(option.Description) == "" || seen[option.ID] {
			return "", fmt.Errorf("decision option %q is empty or duplicated", option.ID)
		}
		seen[option.ID] = true
		fallbackFound = fallbackFound || option.ID == request.Fallback
	}
	if !fallbackFound {
		return "", errors.New("fallback must name one of the declared decision options")
	}
	canonical, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("encode decision request: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}
