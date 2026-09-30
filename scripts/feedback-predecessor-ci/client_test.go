package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGitHubGetRetriesTransientStatusThenSucceeds(t *testing.T) {
	var requests atomic.Int32
	var mu sync.Mutex
	var paths []string
	var authorizations []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mu.Lock()
		paths = append(paths, request.URL.RequestURI())
		authorizations = append(authorizations, request.Header.Get("Authorization"))
		mu.Unlock()
		if requests.Add(1) == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(writer, "private response body")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"value":"ok"}`)
	}))
	defer server.Close()

	var logOutput bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logOutput)
	defer log.SetOutput(previousOutput)

	var result struct {
		Value string `json:"value"`
	}
	err := newGitHubClient(server.URL, "secret-token").getJSON(context.Background(), "/exact/path?ref=abc", &result)
	if err != nil {
		t.Fatalf("getJSON after transient status: %v", err)
	}
	if result.Value != "ok" || requests.Load() != 2 {
		t.Fatalf("result=%+v requests=%d, want value=ok and exactly two requests", result, requests.Load())
	}
	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 || paths[0] != paths[1] || paths[0] != "/exact/path?ref=abc" {
		t.Fatalf("retry changed the exact GET target: %v", paths)
	}
	if len(authorizations) != 2 || authorizations[0] != "Bearer secret-token" || authorizations[1] != authorizations[0] {
		t.Fatalf("retry changed request credentials unexpectedly")
	}
	if !strings.Contains(logOutput.String(), "HTTP status 503") || !strings.Contains(logOutput.String(), "attempt 2/3") {
		t.Fatalf("retry was not observably logged with sanitized status: %q", logOutput.String())
	}
	if strings.Contains(logOutput.String(), server.URL) || strings.Contains(logOutput.String(), "secret-token") || strings.Contains(logOutput.String(), "private response body") {
		t.Fatalf("retry log contains private request or response data: %q", logOutput.String())
	}
}

func TestGitHubGetPersistentTransientFailureIsBoundedAndSanitized(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(writer, "private body that must not be returned")
	}))
	defer server.Close()

	_, err := newGitHubClient(server.URL, "secret-token").get(context.Background(), "/same")
	if err == nil {
		t.Fatal("get succeeded after persistent 502")
	}
	if requests.Load() != maxGetAttempts {
		t.Fatalf("requests=%d, want bounded %d attempts", requests.Load(), maxGetAttempts)
	}
	for _, want := range []string{"status 502", "3 attempt(s)", "statuses [502,502,502]"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not preserve %q", err, want)
		}
	}
	for _, secret := range []string{server.URL, "secret-token", "private body"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error contains private request or response data %q: %v", secret, err)
		}
	}
}

func TestGitHubGetRetriesMixedGatewayTimeoutAndUnavailableStatuses(t *testing.T) {
	statuses := []int{http.StatusGatewayTimeout, http.StatusServiceUnavailable, http.StatusOK}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		index := int(requests.Add(1)) - 1
		if index < len(statuses)-1 {
			writer.WriteHeader(statuses[index])
			return
		}
		writer.WriteHeader(statuses[index])
		_, _ = io.WriteString(writer, "ok")
	}))
	defer server.Close()

	data, err := newGitHubClient(server.URL, "token").get(context.Background(), "/exact")
	if err != nil {
		t.Fatalf("get after 504, 503: %v", err)
	}
	if string(data) != "ok" || requests.Load() != maxGetAttempts {
		t.Fatalf("body=%q requests=%d, want ok after exactly %d attempts", data, requests.Load(), maxGetAttempts)
	}
}

func TestGitHubGetCancellationStopsRetryBackoff(t *testing.T) {
	requestSeen := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestSeen <- struct{}{}
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	retryStarted := make(chan struct{})
	previousOutput := log.Writer()
	log.SetOutput(&signalLogWriter{signal: retryStarted})
	defer log.SetOutput(previousOutput)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := newGitHubClient(server.URL, "token").get(ctx, "/same")
		result <- err
	}()
	<-requestSeen
	select {
	case <-retryStarted:
	case <-time.After(time.Second):
		t.Fatal("first transient response did not enter retry backoff")
	}
	cancel()
	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "status 503") || !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("cancelled backoff error = %v, want preserved 503 and cancellation", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("cancellation did not interrupt retry backoff promptly")
	}
}

func TestWaitBeforeRetryStopsAtDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := waitBeforeRetry(ctx, time.Second)
	if err != context.DeadlineExceeded {
		t.Fatalf("wait error = %v, want context deadline exceeded", err)
	}
}

func TestGitHubGetDoesNotRetryTerminalStatuses(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError, http.StatusTooManyRequests} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				writer.WriteHeader(status)
			}))
			defer server.Close()

			_, err := newGitHubClient(server.URL, "token").get(context.Background(), "/exact")
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("status %d", status)) {
				t.Fatalf("get error=%v, want terminal status %d", err, status)
			}
			if requests.Load() != 1 || !strings.Contains(err.Error(), "1 attempt(s)") {
				t.Fatalf("requests=%d error=%v, want one attempt", requests.Load(), err)
			}
		})
	}
}

func TestGitHubGetPreservesContextErrorsDuringRequestAndBodyRead(t *testing.T) {
	t.Run("canceled while waiting for response headers", func(t *testing.T) {
		requestSeen := make(chan struct{}, 1)
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestSeen <- struct{}{}
			<-request.Context().Done()
		}))
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result := make(chan error, 1)
		go func() {
			_, err := newGitHubClient(server.URL, "secret-token").get(ctx, "/private/path")
			result <- err
		}()
		<-requestSeen
		cancel()
		assertContextError(t, result, context.Canceled, server.URL, "secret-token")
	})

	t.Run("deadline while reading response body", func(t *testing.T) {
		headersSent := make(chan struct{}, 1)
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.(http.Flusher).Flush()
			headersSent <- struct{}{}
			<-request.Context().Done()
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
		defer cancel()
		result := make(chan error, 1)
		go func() {
			_, err := newGitHubClient(server.URL, "secret-token").get(ctx, "/private/path")
			result <- err
		}()
		select {
		case <-headersSent:
		case <-time.After(time.Second):
			t.Fatal("server did not send response headers")
		}
		assertContextError(t, result, context.DeadlineExceeded, server.URL, "secret-token")
	})
}

func assertContextError(t *testing.T, result <-chan error, want error, secrets ...string) {
	t.Helper()
	select {
	case err := <-result:
		if !errors.Is(err, want) {
			t.Fatalf("error = %v, want errors.Is(_, %v)", err, want)
		}
		for _, secret := range secrets {
			if strings.Contains(err.Error(), secret) {
				t.Errorf("error contains private request data %q: %v", secret, err)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("request did not stop promptly after context cancellation")
	}
}

func TestGitHubGetDoesNotRetryMalformedJSON(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(writer, "{")
	}))
	defer server.Close()

	var result map[string]any
	err := newGitHubClient(server.URL, "token").getJSON(context.Background(), "/exact", &result)
	if err == nil || !strings.Contains(err.Error(), "decode GitHub response") {
		t.Fatalf("getJSON error=%v, want decode failure", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests=%d, want one request for malformed JSON", requests.Load())
	}
}

func TestGitHubGetRejectsOversizedBodyInsteadOfTruncating(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		writer.Header().Set("Content-Type", "application/json")
		buffer := bytes.Repeat([]byte{'x'}, 32<<10)
		for written := 0; written < maxResponseBodyBytes+1; written += len(buffer) {
			remaining := maxResponseBodyBytes + 1 - written
			chunk := buffer
			if remaining < len(chunk) {
				chunk = chunk[:remaining]
			}
			if _, err := writer.Write(chunk); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	_, err := newGitHubClient(server.URL, "token").get(context.Background(), "/oversized")
	if err == nil || !strings.Contains(err.Error(), "exceeds 16 MiB limit") {
		t.Fatalf("get error=%v, want fail-closed oversize error", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests=%d, want no retry for oversized response", requests.Load())
	}
}

type signalLogWriter struct {
	signal chan struct{}
	once   sync.Once
}

func (writer *signalLogWriter) Write(data []byte) (int, error) {
	writer.once.Do(func() { close(writer.signal) })
	return len(data), nil
}
