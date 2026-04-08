package amazonpay

import (
    "errors"
)

type Region string

const (
    RegionUS Region = "us"
    RegionEU Region = "eu"
    RegionJP Region = "jp"
)

type Environment string

const (
    EnvironmentSandbox    Environment = "sandbox"
    EnvironmentProduction Environment = "production"
)

type Algorithm string

const (
    AlgorithmAMZNPayRSASSAPSSV2 Algorithm = "AMZN-PAY-RSASSA-PSS-V2"
)

type Config struct {
    PublicKeyID   string
    PrivateKeyPEM []byte
    Region        Region
    Environment   Environment
    Algorithm     Algorithm
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
    if c.Algorithm == "" {
        c.Algorithm = AlgorithmAMZNPayRSASSAPSSV2
    }
    return nil
}

func endpoint(region Region, env Environment) string {
    var host string
    switch region {
    case RegionUS:
        host = "pay-api.amazon.com"
    case RegionEU:
        host = "pay-api.amazon.eu"
    case RegionJP:
        host = "pay-api.amazon.jp"
    default:
        host = "pay-api.amazon.com"
    }

    if env == EnvironmentSandbox {
        return "https://sandbox." + host
    }
    return "https://" + host
}
