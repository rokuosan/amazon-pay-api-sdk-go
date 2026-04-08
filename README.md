# amazon-pay-api-sdk-go

A Go SDK for Amazon Pay API, designed in an idiomatic Go style.

This project is an independent Go port based on the public API surface described by the Node.js SDK documentation.
It does **not** attempt to preserve the original JavaScript interface. Instead, it provides a `Client`-centric API, `context.Context` support, and standard `net/http` integration.

## Status

Initial implementation includes:

- configuration and endpoint resolution
- request signing scaffold for Amazon Pay style signed requests
- low-level `Do` API
- signed header generation
- button signature generation
- high-level resource helpers for common Amazon Pay endpoints

This is an initial cut and may need refinement against live Amazon Pay behavior.

## Install

```bash
go get github.com/rokuosan/amazon-pay-api-sdk-go
```

## Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    amazonpay "github.com/rokuosan/amazon-pay-api-sdk-go"
)

func main() {
    privateKeyPEM, err := os.ReadFile("private.pem")
    if err != nil {
        panic(err)
    }

    client, err := amazonpay.NewClient(amazonpay.Config{
        PublicKeyID: string("PUBLIC_KEY_ID"),
        PrivateKeyPEM: privateKeyPEM,
        Region: amazonpay.RegionJP,
        Environment: amazonpay.EnvironmentSandbox,
        Algorithm: amazonpay.AlgorithmAMZNPayRSASSAPSSV2,
    })
    if err != nil {
        panic(err)
    }

    payload := map[string]any{
        "webCheckoutDetails": map[string]any{
            "checkoutReviewReturnUrl": "https://example.com/review",
            "checkoutResultReturnUrl": "https://example.com/result",
        },
        "storeId": "amzn1.application-oa2-client.xxxxx",
    }

    resp, err := client.CreateCheckoutSession(context.Background(), payload, amazonpay.WithIdempotencyKey("example-idempotency-key"))
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.StatusCode)
    fmt.Println(string(resp.Body))
}
```

## Design Notes

- High-level methods return `*Response` with raw response bytes.
- Callers can decode response JSON into their own structs.
- Request payloads accept `any` and are encoded as JSON.
- Header customization is provided via request options.
- The SDK exposes lower-level signing helpers for advanced use cases.

## Implemented Endpoints

### Checkout v2

- `GetBuyer`
- `CreateCheckoutSession`
- `GetCheckoutSession`
- `UpdateCheckoutSession`
- `CompleteCheckoutSession`
- `GetChargePermission`
- `UpdateChargePermission`
- `CloseChargePermission`
- `CreateCharge`
- `GetCharge`
- `CaptureCharge`
- `CancelCharge`
- `CreateRefund`
- `GetRefund`

### Other APIs

- `DeliveryTrackers`
- `GetAuthorizationToken`
- `InStoreMerchantScan`
- `InStoreCharge`
- `InStoreRefund`
- reporting API helpers

## Caveat

Amazon Pay request signing details are security-sensitive and protocol-specific. This implementation follows the public Node SDK documentation at a high level, but should be validated against Amazon Pay integration tests before production use.
