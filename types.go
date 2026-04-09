package amazonpay

// Price represents an amount and currency pair used by Amazon Pay APIs.
type Price struct {
	Amount       string `json:"amount,omitempty"`
	CurrencyCode string `json:"currencyCode,omitempty"`
}

// ProviderMetadata identifies PSP/solution metadata used in a checkout session.
type ProviderMetadata struct {
	ProviderReferenceID string `json:"providerReferenceId,omitempty"`
}

// WebCheckoutDetails contains return URLs and checkout flow options.
type WebCheckoutDetails struct {
	CheckoutReviewReturnURL string `json:"checkoutReviewReturnUrl,omitempty"`
	CheckoutResultReturnURL string `json:"checkoutResultReturnUrl,omitempty"`
	CheckoutCancelURL       string `json:"checkoutCancelUrl,omitempty"`
	AmazonPayRedirectURL    string `json:"amazonPayRedirectUrl,omitempty"`
}

// AddressDetails contains buyer address information.
type AddressDetails struct {
	Name          string `json:"name,omitempty"`
	AddressLine1  string `json:"addressLine1,omitempty"`
	AddressLine2  string `json:"addressLine2,omitempty"`
	AddressLine3  string `json:"addressLine3,omitempty"`
	City          string `json:"city,omitempty"`
	County        string `json:"county,omitempty"`
	District      string `json:"district,omitempty"`
	StateOrRegion string `json:"stateOrRegion,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
	CountryCode   string `json:"countryCode,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty"`
}

// Buyer contains basic buyer profile information.
type Buyer struct {
	BuyerID              string   `json:"buyerId,omitempty"`
	Name                 string   `json:"name,omitempty"`
	Email                string   `json:"email,omitempty"`
	PhoneNumber          string   `json:"phoneNumber,omitempty"`
	PrimeMembershipTypes []string `json:"primeMembershipTypes,omitempty"`
}

// PaymentIntent represents intent values for checkout/charge flows.
type PaymentIntent string

const (
	PaymentIntentAuthorize            PaymentIntent = "Authorize"
	PaymentIntentAuthorizeWithCapture PaymentIntent = "AuthorizeWithCapture"
	PaymentIntentConfirm              PaymentIntent = "Confirm"
)

// ChargePermissionType indicates whether recurring permission is used.
type ChargePermissionType string

const (
	ChargePermissionTypeOneTime   ChargePermissionType = "OneTime"
	ChargePermissionTypeRecurring ChargePermissionType = "Recurring"
)

// RecurringMetadata specifies recurring payment schedule metadata.
type RecurringMetadata struct {
	Frequency *UnitFrequency `json:"frequency,omitempty"`
	Amount    *Price         `json:"amount,omitempty"`
}

// UnitFrequency represents recurring interval values.
type UnitFrequency struct {
	Unit  string `json:"unit,omitempty"`
	Value string `json:"value,omitempty"`
}

// PaymentDetails includes amount and payment behavior details.
type PaymentDetails struct {
	ChargeAmount                  *Price `json:"chargeAmount,omitempty"`
	TotalOrderAmount              *Price `json:"totalOrderAmount,omitempty"`
	SoftDescriptor                string `json:"softDescriptor,omitempty"`
	AllowOvercharge               bool   `json:"allowOvercharge,omitempty"`
	CanHandlePendingAuthorization bool   `json:"canHandlePendingAuthorization,omitempty"`
}

// CreateCheckoutSessionRequest maps checkout session creation payload.
type CreateCheckoutSessionRequest struct {
	WebCheckoutDetails   *WebCheckoutDetails  `json:"webCheckoutDetails,omitempty"`
	StoreID              string               `json:"storeId,omitempty"`
	MerchantMetadata     map[string]string    `json:"merchantMetadata,omitempty"`
	PaymentDetails       *PaymentDetails      `json:"paymentDetails,omitempty"`
	ChargePermissionType ChargePermissionType `json:"chargePermissionType,omitempty"`
	RecurringMetadata    *RecurringMetadata   `json:"recurringMetadata,omitempty"`
	PlatformID           string               `json:"platformId,omitempty"`
	ProviderMetadata     *ProviderMetadata    `json:"providerMetadata,omitempty"`
}

// UpdateCheckoutSessionRequest maps checkout session update payload.
type UpdateCheckoutSessionRequest struct {
	WebCheckoutDetails *WebCheckoutDetails `json:"webCheckoutDetails,omitempty"`
	PaymentDetails     *PaymentDetails     `json:"paymentDetails,omitempty"`
	MerchantMetadata   map[string]string   `json:"merchantMetadata,omitempty"`
}

// CompleteCheckoutSessionRequest maps checkout session completion payload.
type CompleteCheckoutSessionRequest struct {
	ChargeAmount *Price `json:"chargeAmount,omitempty"`
}

// UpdateChargePermissionRequest maps charge permission update payload.
type UpdateChargePermissionRequest struct {
	MerchantMetadata map[string]string `json:"merchantMetadata,omitempty"`
}

// CloseChargePermissionRequest maps charge permission close payload.
type CloseChargePermissionRequest struct {
	ClosureReason        string `json:"closureReason,omitempty"`
	CancelPendingCharges bool   `json:"cancelPendingCharges,omitempty"`
}

// CreateChargeRequest maps charge creation payload.
type CreateChargeRequest struct {
	ChargePermissionID            string `json:"chargePermissionId,omitempty"`
	ChargeAmount                  *Price `json:"chargeAmount,omitempty"`
	CaptureNow                    bool   `json:"captureNow,omitempty"`
	SoftDescriptor                string `json:"softDescriptor,omitempty"`
	CanHandlePendingAuthorization bool   `json:"canHandlePendingAuthorization,omitempty"`
}

// CaptureChargeRequest maps capture payload for an authorized charge.
type CaptureChargeRequest struct {
	CaptureAmount  *Price `json:"captureAmount,omitempty"`
	SoftDescriptor string `json:"softDescriptor,omitempty"`
}

// CancelChargeRequest maps cancel payload for a charge.
type CancelChargeRequest struct {
	CancellationReason string `json:"cancellationReason,omitempty"`
}

// CreateRefundRequest maps refund creation payload.
type CreateRefundRequest struct {
	ChargeID       string `json:"chargeId,omitempty"`
	RefundAmount   *Price `json:"refundAmount,omitempty"`
	SoftDescriptor string `json:"softDescriptor,omitempty"`
}

// DeliveryTrackerDetails represents carrier/tracking info.
type DeliveryTrackerDetails struct {
	CarrierCode    string `json:"carrierCode,omitempty"`
	TrackingNumber string `json:"trackingNumber,omitempty"`
}

// DeliveryTrackersRequest maps deliveryTrackers API payload.
type DeliveryTrackersRequest struct {
	ChargeID               string                  `json:"chargeId,omitempty"`
	DeliveryTrackerDetails *DeliveryTrackerDetails `json:"deliveryTrackerDetails,omitempty"`
}

// MerchantScanRequest maps in-store merchant scan payload.
type MerchantScanRequest struct {
	ScanData string `json:"scanData,omitempty"`
}

// InStoreChargeRequest maps in-store charge payload.
type InStoreChargeRequest struct {
	ScanData       string `json:"scanData,omitempty"`
	ChargeAmount   *Price `json:"chargeAmount,omitempty"`
	SoftDescriptor string `json:"softDescriptor,omitempty"`
}

// InStoreRefundRequest maps in-store refund payload.
type InStoreRefundRequest struct {
	ChargeID     string `json:"chargeId,omitempty"`
	RefundAmount *Price `json:"refundAmount,omitempty"`
}
