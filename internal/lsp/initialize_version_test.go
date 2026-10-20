package lsp

import (
	"encoding/json"
	"testing"
)

func TestInitializeAdvertisesBoundOrExplicitFallbackVersion(t *testing.T) {
	for _, tc := range []struct {
		name      string
		newServer func() *Server
		want      string
	}{
		{
			name:      "bound executable",
			newServer: func() *Server { return NewServerWithVersion("0.6.37-dev") },
			want:      "0.6.37-dev",
		},
		{name: "direct library default", newServer: func() *Server { return NewServer() }, want: "unspecified"},
		{name: "empty metadata", newServer: func() *Server { return NewServerWithVersion("") }, want: "unspecified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.newServer()
			response, notifications, err := server.initialize(requestEnvelope{ID: json.RawMessage("1")})
			if err != nil || len(notifications) != 0 || response == nil || response.Error != nil {
				t.Fatalf("initialize = response %#v, notifications %d, error %v", response, len(notifications), err)
			}
			var result InitializeResult
			if err := json.Unmarshal(response.Result, &result); err != nil {
				t.Fatal(err)
			}
			if result.ServerInfo.Name != "gooo-lsp" || result.ServerInfo.Version != tc.want {
				t.Fatalf("server info = %#v, want version %q", result.ServerInfo, tc.want)
			}
		})
	}
}
