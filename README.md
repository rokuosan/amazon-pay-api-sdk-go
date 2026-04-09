# amazon-pay-api-sdk-go

Amazon Pay API向けのGo SDKです。Goらしい`Client`中心の設計で、`context.Context`、`net/http`、および署名付きリクエスト生成を提供します。

## 現在の実装内容

このSDKには以下が実装されています。

- `Config`バリデーションとリージョン別エンドポイント解決
- Amazon Pay署名ヘッダー生成（`SignHeaders` / `GetSignedHeaders`）
- ボタンペイロード署名生成（`GenerateButtonSignature`）
- 低レベルAPI呼び出し（`APICall`）
- 高レベルAPIメソッド群（Checkout / Charge / Refund / Reports / Disputes / In-Storeなど）
- HTTPステータスに応じたリトライ（429、5xxなど）

## インストール

```bash
go get github.com/rokuosan/amazon-pay-api-sdk-go
```

## クイックスタート

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

## 設計メモ

- 高レベルメソッドは`*Response`を返し、レスポンスボディは`[]byte`で取得できます。
- リクエストpayloadは`any`を受け取り、`string`/`[]byte`以外はJSONエンコードされます。
- 追加ヘッダーは各メソッドの`headers map[string]string`で渡せます。
- `APICall`で低レベルに`Method`/`Path`/`QueryParams`を直接指定できます。

## コンフィグ (`Config`)

`NewClient`に渡す主な項目:

- `PublicKeyID`（必須）
- `PrivateKeyPEM`（必須）
- `Region`（必須: `na/us/eu/de/uk/jp`）
- `Environment`（`sandbox` or `live`）
- `Algorithm`（`AMZN-PAY-RSASSA-PSS` or `AMZN-PAY-RSASSA-PSS-V2`）
- `OverrideServiceURL`（テスト/モック向け）
- `HTTPClient`（未指定時は`http.DefaultClient`）
- `MaxRetries`（未指定時は`3`）
- `Now`（時刻注入、テスト向け）
- `UserAgent`（未指定時はSDK標準値）

## 実装済みメソッド

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

## エラーハンドリング

- HTTPステータスが`>= 400`の場合は`*HTTPError`が返ります。
- `AsHTTPError`ヘルパーでステータスコードやレスポンスボディを取得できます。

## 注意事項

署名やAPIの詳細仕様はセキュリティ上重要です。本SDKは公開情報をもとに実装されていますが、本番利用前にAmazon Payの最新公式仕様と突き合わせて検証してください。
