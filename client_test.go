package amazonpay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testPrivateKeyPEM(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func TestAPICall_SignsAndSendsRequest(t *testing.T) {
	var gotPath, gotAuth, gotBody, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("authorization")
		gotBody = readBody(t, r)
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		PublicKeyID:        "LIVE-XXXX",
		PrivateKeyPEM:      testPrivateKeyPEM(t),
		Region:             RegionJP,
		Environment:        EnvironmentSandbox,
		Algorithm:          AlgorithmAMZNPayRSASSAPSSV2,
		OverrideServiceURL: srv.URL,
		Now:                func() time.Time { return time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resp, err := c.APICall(context.Background(), APIRequest{
		Method:  http.MethodPost,
		Path:    "checkoutSessions",
		Payload: map[string]any{"a": "b"},
		QueryParams: map[string]string{
			"z": "2",
			"a": "1",
		},
	})
	if err != nil {
		t.Fatalf("APICall: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if gotPath != "/v2/checkoutSessions" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery != "a=1&z=2" {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
	if gotAuth == "" {
		t.Fatal("authorization header missing")
	}
	if gotBody != `{"a":"b"}` {
		t.Fatalf("unexpected body: %s", gotBody)
	}
}

func TestResolvePath_WithNonEnvSpecificKey(t *testing.T) {
	c, err := NewClient(Config{PublicKeyID: "ABC", PrivateKeyPEM: testPrivateKeyPEM(t), Region: RegionUS, Environment: EnvironmentSandbox})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.resolvePath("checkoutSessions"); got != "sandbox/v2/checkoutSessions" {
		t.Fatalf("unexpected path: %s", got)
	}
}

func TestAPICall_RetryOnTooManyRequests(t *testing.T) {
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`rate limited`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{PublicKeyID: "LIVE-KEY", PrivateKeyPEM: testPrivateKeyPEM(t), Region: RegionUS, OverrideServiceURL: srv.URL, MaxRetries: 1})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	resp, err := c.APICall(context.Background(), APIRequest{Method: http.MethodGet, Path: "buyers/token"})
	if err != nil {
		t.Fatalf("APICall: %v", err)
	}
	if resp.StatusCode != http.StatusOK || count != 2 {
		t.Fatalf("status=%d count=%d", resp.StatusCode, count)
	}
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}
