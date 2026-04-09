package amazonpay

import (
	"strings"
	"testing"
	"time"
)

func TestSignHeaders_ContainsExpectedHeaders(t *testing.T) {
	c, err := NewClient(Config{
		PublicKeyID:   "LIVE-KEY",
		PrivateKeyPEM: testPrivateKeyPEM(t),
		Region:        RegionDE,
		Algorithm:     AlgorithmAMZNPayRSASSAPSSV2,
		Now:           func() time.Time { return time.Date(2025, 8, 20, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	headers, err := c.SignHeaders(APIRequest{Method: "POST", Path: "v2/checkoutSessions", Headers: map[string]string{"x-custom": "1"}}, "{}")
	if err != nil {
		t.Fatalf("SignHeaders: %v", err)
	}
	if headers["x-amz-pay-region"] != "eu" {
		t.Fatalf("unexpected region: %s", headers["x-amz-pay-region"])
	}
	if !strings.Contains(headers["authorization"], "AMZN-PAY-RSASSA-PSS-V2 PublicKeyId=LIVE-KEY") {
		t.Fatalf("unexpected auth header: %s", headers["authorization"])
	}
	if headers["x-amz-pay-date"] != "20250820T120000Z" {
		t.Fatalf("unexpected x-amz-pay-date: %s", headers["x-amz-pay-date"])
	}
	if !strings.Contains(headers["authorization"], "SignedHeaders=accept;content-type;user-agent;x-amz-pay-date;x-amz-pay-host;x-amz-pay-region;x-custom") {
		t.Fatalf("unexpected signed headers in authorization: %s", headers["authorization"])
	}
}

func TestGenerateButtonSignature(t *testing.T) {
	c, err := NewClient(Config{PublicKeyID: "LIVE-KEY", PrivateKeyPEM: testPrivateKeyPEM(t), Region: RegionUS})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	sig, err := c.GenerateButtonSignature("{\"hello\":\"world\"}")
	if err != nil {
		t.Fatalf("GenerateButtonSignature: %v", err)
	}
	if sig == "" {
		t.Fatal("signature should not be empty")
	}
}

func TestGenerateButtonSignature_AcceptsStructuredPayload(t *testing.T) {
	c, err := NewClient(Config{PublicKeyID: "LIVE-KEY", PrivateKeyPEM: testPrivateKeyPEM(t), Region: RegionUS})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.GenerateButtonSignature(map[string]any{"a": "b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignHeaders_UsesOnlyHostForOverrideServiceURL(t *testing.T) {
	c, err := NewClient(Config{
		PublicKeyID:        "LIVE-KEY",
		PrivateKeyPEM:      testPrivateKeyPEM(t),
		Region:             RegionUS,
		OverrideServiceURL: "https://localhost:8443/some/prefix",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	headers, err := c.SignHeaders(APIRequest{Method: "GET", Path: "v2/reports"}, "")
	if err != nil {
		t.Fatalf("SignHeaders: %v", err)
	}
	if headers["x-amz-pay-host"] != "localhost:8443" {
		t.Fatalf("unexpected host: %s", headers["x-amz-pay-host"])
	}
}
