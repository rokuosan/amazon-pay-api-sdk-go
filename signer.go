package amazonpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type HTTPError struct {
	StatusCode int
	Header     map[string][]string
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("amazonpay: http status %d", e.StatusCode)
}

func AsHTTPError(err error, target **HTTPError) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		*target = httpErr
		return true
	}
	return false
}

func (c *Client) SignHeaders(req APIRequest, payload string) (map[string]string, error) {
	host, _ := endpointHost(c.cfg.Region)
	if c.cfg.OverrideServiceURL != "" {
		host = strings.TrimPrefix(strings.TrimPrefix(c.cfg.OverrideServiceURL, "https://"), "http://")
		host = strings.TrimSuffix(host, "/")
	}

	headers := map[string]string{}
	for k, v := range req.Headers {
		headers[k] = v
	}
	headers["x-amz-pay-region"] = normalizeRegion(c.cfg.Region)
	headers["x-amz-pay-host"] = host
	headers["x-amz-pay-date"] = c.cfg.Now().UTC().Format("2006-01-02T15:04:05Z")
	headers["content-type"] = "application/json"
	headers["accept"] = "application/json"
	headers["user-agent"] = c.cfg.UserAgent

	sorted := sortedHeaderKeys(headers)
	signedHeaders := strings.Join(sorted, ";")

	canonical := req.Method + "\n/" + strings.TrimLeft(req.Path, "/") + "\n" + encodeQuery(req.QueryParams) + "\n"
	for _, k := range sorted {
		canonical += strings.ToLower(k) + ":" + headers[k] + "\n"
	}
	canonical += "\n" + signedHeaders + "\n" + sha256Hex(payload)

	alg, saltLen, err := resolveAlgorithm(c.cfg.Algorithm)
	if err != nil {
		return nil, err
	}
	stringToSign := alg + "\n" + sha256Hex(canonical)
	sig, err := c.sign(stringToSign, saltLen)
	if err != nil {
		return nil, err
	}

	headers["authorization"] = fmt.Sprintf("%s PublicKeyId=%s, SignedHeaders=%s, Signature=%s", alg, c.cfg.PublicKeyID, signedHeaders, sig)
	return headers, nil
}

func (c *Client) GenerateButtonSignature(payload any) (string, error) {
	var payloadText string
	switch p := payload.(type) {
	case string:
		payloadText = p
	case []byte:
		payloadText = string(p)
	default:
		return "", errors.New("amazonpay: GenerateButtonSignature payload must be string or []byte")
	}
	alg, saltLen, err := resolveAlgorithm(c.cfg.Algorithm)
	if err != nil {
		return "", err
	}
	return c.sign(alg+"\n"+sha256Hex(payloadText), saltLen)
}

func (c *Client) sign(message string, saltLen int) (string, error) {
	priv, err := parsePrivateKey(c.cfg.PrivateKeyPEM)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hash[:], &rsa.PSSOptions{SaltLength: saltLen, Hash: crypto.SHA256})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func parsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("amazonpay: invalid PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("amazonpay: parse private key: %w", err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("amazonpay: private key is not RSA")
	}
	return rsaKey, nil
}

func normalizeRegion(region Region) string {
	switch strings.ToLower(string(region)) {
	case "us", "na":
		return "na"
	case "eu", "de", "uk":
		return "eu"
	case "jp":
		return "jp"
	default:
		return strings.ToLower(string(region))
	}
}

func sortedHeaderKeys(headers map[string]string) []string {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})
	return keys
}

func sha256Hex(in string) string {
	h := sha256.Sum256([]byte(in))
	return hex.EncodeToString(h[:])
}

func resolveAlgorithm(alg Algorithm) (name string, saltLen int, err error) {
	switch alg {
	case "", AlgorithmAMZNPayRSASSAPSS:
		return string(AlgorithmAMZNPayRSASSAPSS), 20, nil
	case AlgorithmAMZNPayRSASSAPSSV2:
		return string(AlgorithmAMZNPayRSASSAPSSV2), 32, nil
	default:
		return "", 0, fmt.Errorf("amazonpay: unsupported algorithm %q", alg)
	}
}
