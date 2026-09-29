package decisionroute

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	FallbackNotConfigured = "NOT_CONFIGURED"
	FallbackUnavailable   = "PROVIDER_UNAVAILABLE"
	FallbackHTTPError     = "PROVIDER_HTTP_ERROR"
	FallbackInvalidResult = "PROVIDER_RESULT_INVALID"
)

type layaRequest struct {
	State     map[string]string             `json:"state"`
	Questions map[string]layaChoiceQuestion `json:"questions"`
}

type layaChoiceQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type layaResponse struct {
	Model   string         `json:"model"`
	Routing map[string]any `json:"routing"`
	Answers map[string]struct {
		Choice           string             `json:"choice"`
		Probabilities    map[string]float64 `json:"probabilities"`
		Confidence       *float64           `json:"confidence"`
		AnswerConfidence *float64           `json:"answer_confidence"`
	} `json:"answers"`
}

func Resolve(ctx context.Context, request Request, endpoint, apiKey string) (Receipt, error) {
	digest, err := Validate(request)
	if err != nil {
		return Receipt{}, err
	}
	if strings.TrimSpace(endpoint) == "" {
		return fallback(request, digest, FallbackNotConfigured), nil
	}
	if err := validEndpoint(endpoint); err != nil {
		return fallback(request, digest, FallbackUnavailable), nil
	}
	payload := buildLayaRequest(request)
	body, _ := json.Marshal(payload)
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fallback(request, digest, FallbackUnavailable), nil
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{Timeout: 90 * time.Second}
	response, err := client.Do(httpRequest)
	if err != nil {
		return fallback(request, digest, FallbackUnavailable), nil
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fallback(request, digest, FallbackHTTPError), nil
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20+1))
	if err != nil || len(raw) > 1<<20 {
		return fallback(request, digest, FallbackInvalidResult), nil
	}
	var result layaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return fallback(request, digest, FallbackInvalidResult), nil
	}
	answer, ok := result.Answers[request.Question.ID]
	if !ok || !validChoice(request, answer.Choice) || !validProbabilities(request, answer.Probabilities) || !validConfidence(answer.Confidence) || !validConfidence(answer.AnswerConfidence) {
		return fallback(request, digest, FallbackInvalidResult), nil
	}
	return Receipt{
		Schema: ReceiptSchema, Mode: "laya", Selected: answer.Choice, Provider: "laya",
		Model: result.Model, ModelRevision: readModelRevision(ctx, endpoint, result.Routing),
		Routing: result.Routing, Probabilities: answer.Probabilities,
		Confidence: answer.Confidence, AnswerConfidence: answer.AnswerConfidence,
		RequestSHA256: digest,
	}, nil
}

func readModelRevision(ctx context.Context, endpoint string, routing map[string]any) string {
	parsed, err := url.Parse(endpoint)
	if err != nil || !strings.HasSuffix(parsed.Path, "/v1/systemone") {
		return ""
	}
	model, _ := routing["model"].(string)
	if model == "" {
		return ""
	}
	parsed.Path = strings.TrimSuffix(parsed.Path, "/v1/systemone") + "/health"
	parsed.RawQuery, parsed.Fragment = "", ""
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return ""
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return ""
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ""
	}
	var health struct {
		Revisions map[string]string `json:"revisions"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 1<<16)).Decode(&health) != nil {
		return ""
	}
	revision := health.Revisions[model]
	decoded, err := hex.DecodeString(revision)
	if err != nil || len(decoded) != 20 {
		return ""
	}
	return revision
}

func validEndpoint(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil {
		return errors.New("Laya URL must be an http(s) URL without embedded credentials")
	}
	return nil
}

func buildLayaRequest(request Request) layaRequest {
	criteria := make(map[string]string, len(request.Question.Options))
	for _, option := range request.Question.Options {
		criteria[option.ID] = option.Description
	}
	return layaRequest{
		State: map[string]string{"request": request.State},
		Questions: map[string]layaChoiceQuestion{request.Question.ID: {
			Type: "choice", Instructions: request.Question.Instructions, Criteria: criteria,
		}},
	}
}

func validChoice(request Request, selected string) bool {
	for _, option := range request.Question.Options {
		if option.ID == selected {
			return true
		}
	}
	return false
}

func validProbabilities(request Request, values map[string]float64) bool {
	if len(values) == 0 {
		return true
	}
	for key, value := range values {
		if !validChoice(request, key) || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return false
		}
	}
	return true
}

func validConfidence(value *float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= 0 && *value <= 1)
}

func fallback(request Request, digest, reason string) Receipt {
	return Receipt{
		Schema: ReceiptSchema, Mode: "deterministic_fallback", Selected: request.Fallback,
		FallbackReason: reason, Provider: "deterministic", RequestSHA256: digest,
	}
}
