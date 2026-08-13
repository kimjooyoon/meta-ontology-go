package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func parseFailure(id json.RawMessage, ctx context.Context, err error) *responseEnvelope {
	if ctx.Err() != nil {
		return nil
	}
	return responseOrNil(id, internalError, err.Error())
}

func errIfCanceled(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func decodeParams(raw json.RawMessage, target any) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func responseOrNil(id json.RawMessage, code int, message string) *responseEnvelope {
	if id == nil {
		return nil
	}
	return errorResponse(id, code, message)
}

func resultResponse(id json.RawMessage, result any) *responseEnvelope {
	encoded, err := json.Marshal(result)
	if err != nil {
		return errorResponse(id, internalError, err.Error())
	}
	return &responseEnvelope{JSONRPC: jsonRPCVersion, ID: responseID(id), Result: encoded}
}

func errorResponse(id json.RawMessage, code int, message string) *responseEnvelope {
	return &responseEnvelope{JSONRPC: jsonRPCVersion, ID: responseID(id), Error: &errorObject{Code: code, Message: message}}
}

func responseID(id json.RawMessage) json.RawMessage {
	if id == nil {
		return json.RawMessage("null")
	}
	return id
}

func writeResponse(output io.Writer, response *responseEnvelope) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("lsp: encode response: %w", err)
	}
	return WriteMessage(output, payload)
}

func oneNotification(notification []byte, err error) [][]byte {
	if err != nil || notification == nil {
		return nil
	}
	return [][]byte{notification}
}
