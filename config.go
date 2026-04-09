package amazonpay

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	sdkVersion = "0.1.0"
	apiVersion = "v2"
)

type Region string

const (
	RegionNA Region = "na"
	RegionUS Region = "us"
	RegionEU Region = "eu"
	RegionDE Region = "de"
	RegionUK Region = "uk"
	RegionJP Region = "jp"
)

type Environment string

const (
	EnvironmentSandbox Environment = "sandbox"
	EnvironmentLive    Environment = "live"
)

type Algorithm string

const (
	AlgorithmAMZNPayRSASSAPSS   Algorithm = "AMZN-PAY-RSASSA-PSS"
	AlgorithmAMZNPayRSASSAPSSV2 Algorithm = "AMZN-PAY-RSASSA-PSS-V2"
)

type Config struct {
	PublicKeyID        string
	PrivateKeyPEM      []byte
	Region             Region
	Environment        Environment
	Algorithm          Algorithm
	OverrideServiceURL string
	HTTPClient         *http.Client
	MaxRetries         int
	Now                func() time.Time
	UserAgent          string
}

func (c Config) validate() error {
	if c.PublicKeyID == "" {
		return errors.New("amazonpay: PublicKeyID is required")
	}
	if len(c.PrivateKeyPEM) == 0 {
		return errors.New("amazonpay: PrivateKeyPEM is required")
	}
	if c.Region == "" {
		return errors.New("amazonpay: Region is required")
	}
	if _, ok := endpointHost(c.Region); !ok {
		return fmt.Errorf("amazonpay: unsupported region %q", c.Region)
	}
	if c.Algorithm == "" {
		c.Algorithm = AlgorithmAMZNPayRSASSAPSS
	}
	if c.MaxRetries < 0 {
		return errors.New("amazonpay: MaxRetries must be >= 0")
	}
	return nil
}

func endpointHost(region Region) (string, bool) {
	switch strings.ToLower(string(region)) {
	case "na", "us":
		return "pay-api.amazon.com", true
	case "eu", "de", "uk":
		return "pay-api.amazon.eu", true
	case "jp":
		return "pay-api.amazon.jp", true
	default:
		return "", false
	}
}

func isEnvSpecificPublicKeyID(publicKeyID string) bool {
	upper := strings.ToUpper(publicKeyID)
	return strings.HasPrefix(upper, "LIVE") || strings.HasPrefix(upper, "SANDBOX")
}
