package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"runtime"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const documentCacheContract = "gooo-lsp-document-cache:v1;reuse=exact-input;invalidation=source-profile-toolchain-contract"

type documentCacheIdentity struct {
	profileDigest   string
	toolchainDigest string
	contractDigest  string
}

type documentCacheKey struct {
	sourceDigest    string
	profileDigest   string
	toolchainDigest string
	contractDigest  string
}

func newDocumentCacheIdentity(parser Parser) documentCacheIdentity {
	support := syntax.CurrentEntityFieldsSupport()
	profile := fmt.Sprintf("%s|%d|%s|%s", support.State, support.Profile.Version, support.Profile.ID, support.Profile.Digest)
	toolchain := runtime.Version() + "|" + runtime.GOOS + "|" + runtime.GOARCH
	contract := documentCacheContract + "|parser=" + reflect.TypeOf(parser).String()
	return documentCacheIdentity{
		profileDigest:   digestText(profile),
		toolchainDigest: digestText(toolchain),
		contractDigest:  digestText(contract),
	}
}

func (server *Server) cacheKey(source string) documentCacheKey {
	return documentCacheKey{
		sourceDigest:    digestText(source),
		profileDigest:   server.cacheIdentity.profileDigest,
		toolchainDigest: server.cacheIdentity.toolchainDigest,
		contractDigest:  server.cacheIdentity.contractDigest,
	}
}

func digestText(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func (server *Server) lifecycleState() (shutdown, exited bool) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return server.shutdown, server.exited
}

func (server *Server) isInitialized() bool {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return server.initialized
}

func (server *Server) markExited() {
	server.mu.Lock()
	server.exited = true
	server.mu.Unlock()
}
