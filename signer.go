package amazonpay

import (
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "encoding/base64"
    "encoding/pem"
    "fmt"
    "net/http"
    "time"
)

func (c *Client) sign(req *http.Request, payload any) error {
    // NOTE: This is a simplified scaffold. Canonical request construction must be aligned with Amazon Pay spec.

    timestamp := time.Now().UTC().Format(time.RFC3339)
    req.Header.Set("x-amz-pay-date", timestamp)
    req.Header.Set("x-amz-pay-region", string(c.cfg.Region))

    canonical := req.Method + "\n" + req.URL.Path

    hash := sha256.Sum256([]byte(canonical))

    priv, err := parsePrivateKey(c.cfg.PrivateKeyPEM)
    if err != nil {
        return err
    }

    sig, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hash[:], nil)
    if err != nil {
        return err
    }

    signature := base64.StdEncoding.EncodeToString(sig)

    req.Header.Set("Authorization", fmt.Sprintf("%s:%s", c.cfg.PublicKeyID, signature))

    return nil
}

func parsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
    block, _ := pem.Decode(pemBytes)
    if block == nil {
        return nil, fmt.Errorf("invalid PEM")
    }

    key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err == nil {
        return key, nil
    }

    parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
        return nil, err
    }

    rsaKey, ok := parsed.(*rsa.PrivateKey)
    if !ok {
        return nil, fmt.Errorf("not RSA key")
    }
    return rsaKey, nil
}

func (c *Client) GenerateButtonSignature(payload []byte) (string, error) {
    hash := sha256.Sum256(payload)

    priv, err := parsePrivateKey(c.cfg.PrivateKeyPEM)
    if err != nil {
        return "", err
    }

    sig, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hash[:], nil)
    if err != nil {
        return "", err
    }

    return base64.StdEncoding.EncodeToString(sig), nil
}
