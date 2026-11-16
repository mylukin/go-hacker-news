package ai

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type deadlineTransport struct {
	t      *testing.T
	called bool
}

func (tr *deadlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tr.called = true
	deadline, ok := req.Context().Deadline()
	if !ok {
		tr.t.Fatal("request has no deadline")
	}
	remaining := time.Until(deadline)
	if remaining < 299*time.Second || remaining > 300*time.Second {
		tr.t.Fatalf("request deadline = %v; want configured 300 seconds", remaining)
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))}, nil
}

func TestGenerateTextUsesConfiguredTimeout(t *testing.T) {
	original := http.DefaultTransport
	transport := &deadlineTransport{t: t}
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = original })
	client := NewOpenAIClient("test-key", "http://example.invalid/v1", "balanced", 16384, 300)
	result, err := client.GenerateText("system", "user")
	if err != nil {
		t.Fatal(err)
	}
	if !transport.called || result != "ok" {
		t.Fatalf("unexpected completion: %q", result)
	}
}
