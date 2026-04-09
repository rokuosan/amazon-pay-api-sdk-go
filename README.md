# amazon-pay-api-sdk-go

A Go SDK for the Amazon Pay API.
It follows an idiomatic Go design with a `Client`-centric interface, `context.Context` support, and `net/http` integration.

## Current Implementation

This SDK currently includes:

- `Config` validation and region-based endpoint resolution
- Amazon Pay signed header generation (`SignHeaders` / `GetSignedHeaders`)
- Button payload signature generation (`GenerateButtonSignature`)
- Low-level request execution via `APICall`
- High-level API helpers (Checkout, Charge, Refund, Reports, Disputes, In-Store, etc.)
- Retry behavior for retryable HTTP statuses (for example, 429 and 5xx)

## Install

```bash
go get github.com/rokuosan/amazon-pay-api-sdk-go
```

## Quick Start

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
        PublicKeyID:   "PUBLIC_KEY_ID",
        PrivateKeyPEM: privateKeyPEM,
        Region:        amazonpay.RegionJP,
        Environment:   amazonpay.EnvironmentSandbox,
        Algorithm:     amazonpay.AlgorithmAMZNPayRSASSAPSSV2,
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

    headers := map[string]string{
        "x-amz-pay-idempotency-key": "example-idempotency-key",
    }

    resp, err := client.CreateCheckoutSession(context.Background(), payload, headers)
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.StatusCode)
    fmt.Println(string(resp.Body))
}
```

## Design Notes

- High-level methods return `*Response` with raw response bytes (`[]byte`).
- Request payloads accept `any`; non-`string` and non-`[]byte` payloads are JSON-encoded.
- Additional headers can be passed using `headers map[string]string`.
- `APICall` is available for low-level control of method/path/query/payload.

## Configuration (`Config`)

Main fields accepted by `NewClient`:

- `PublicKeyID` (required)
- `PrivateKeyPEM` (required)
- `Region` (required: `na/us/eu/de/uk/jp`)
- `Environment` (`sandbox` or `live`)
- `Algorithm` (`AMZN-PAY-RSASSA-PSS` or `AMZN-PAY-RSASSA-PSS-V2`)
- `OverrideServiceURL` (useful for testing/mocking)
- `HTTPClient` (defaults to `http.DefaultClient`)
- `MaxRetries` (defaults to `3`)
- `Now` (time injection, useful for tests)
- `UserAgent` (defaults to an SDK-generated value)

## Implemented Methods

### Checkout / Buyer / Charge Permission / Charge / Refund

- `GetBuyer`
- `CreateCheckoutSession`
- `GetCheckoutSession`
- `UpdateCheckoutSession`
- `CompleteCheckoutSession`
- `FinalizeCheckoutSession`
- `GetChargePermission`
- `UpdateChargePermission`
- `CloseChargePermission`
- `CreateCharge`
- `GetCharge`
- `UpdateCharge`
- `CaptureCharge`
- `CancelCharge`
- `CreateRefund`
- `GetRefund`

### Reporting

- `GetReports`
- `GetReportByID`
- `GetReportDocument`
- `GetReportSchedules`
- `GetReportScheduleByID`
- `CreateReport`
- `CreateReportSchedule`
- `CancelReportSchedule`

### In-Store / Authorization / Delivery

- `GetAuthorizationToken`
- `DeliveryTrackers`
- `MerchantScan`
- `InStoreCharge`
- `InStoreRefund`

### Merchant Account

- `CreateMerchantAccount`
- `UpdateMerchantAccount`
- `DeleteMerchantAccount`
- `MerchantAccountClaim`

### Disputes / Files

- `CreateDispute`
- `GetDispute`
- `UpdateDispute`
- `ContestDispute`
- `UploadFile`

## Error Handling

- HTTP responses with status `>= 400` return `*HTTPError`.
- Use `AsHTTPError` to inspect status code, headers, and response body.

## Caveat

Request signing and API semantics are security-sensitive.
Before production use, validate behavior against the latest official Amazon Pay specifications and integration tests.
