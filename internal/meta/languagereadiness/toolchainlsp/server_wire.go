package toolchainlsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcMessage struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

func appendRPC(target *bytes.Buffer, id *int, method string, params any) error {
	payload := map[string]any{"jsonrpc": "2.0", "method": method}
	if id != nil {
		payload["id"] = *id
	}
	if params != nil {
		payload["params"] = params
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return lsp.WriteMessage(target, raw)
}

func readRPC(raw []byte) ([]rpcMessage, error) {
	reader := bytes.NewReader(raw)
	messages := make([]rpcMessage, 0)
	for {
		payload, err := lsp.ReadMessage(reader)
		if errors.Is(err, io.EOF) {
			return messages, nil
		}
		if err != nil {
			return nil, err
		}
		var message rpcMessage
		if err := json.Unmarshal(payload, &message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
}
