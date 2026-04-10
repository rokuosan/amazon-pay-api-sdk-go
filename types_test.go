package amazonpay

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreateCheckoutSessionRequest_OmitsEmptyNestedObjects(t *testing.T) {
	payload := CreateCheckoutSessionRequest{StoreID: "store-id"}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	got := string(b)
	if got != `{"storeId":"store-id"}` {
		t.Fatalf("unexpected json: %s", got)
	}
}

func TestCreateCheckoutSessionRequest_IncludesNestedObjectsWhenSet(t *testing.T) {
	payload := CreateCheckoutSessionRequest{
		StoreID: "store-id",
		WebCheckoutDetails: &WebCheckoutDetails{
			CheckoutReviewReturnURL: "https://example.com/review",
		},
		PaymentDetails: &PaymentDetails{
			ChargeAmount: &Price{Amount: "100", CurrencyCode: "JPY"},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	got := string(b)
	if !containsAll(got,
		`"storeId":"store-id"`,
		`"webCheckoutDetails":{"checkoutReviewReturnUrl":"https://example.com/review"}`,
		`"paymentDetails":{"chargeAmount":{"amount":"100","currencyCode":"JPY"}}`,
	) {
		t.Fatalf("unexpected json: %s", got)
	}
}

func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
