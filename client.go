package amazonpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type APIRequest struct {
	Method      string
	Path        string
	Payload     any
	Headers     map[string]string
	QueryParams map[string]string
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Algorithm == "" {
		cfg.Algorithm = AlgorithmAMZNPayRSASSAPSS
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = fmt.Sprintf("amazon-pay-api-sdk-go/%s (Go/%s; %s)", sdkVersion, runtime.Version(), runtime.GOOS)
	}
	return &Client{cfg: cfg, httpClient: cfg.HTTPClient}, nil
}

func (c *Client) APICall(ctx context.Context, req APIRequest) (*Response, error) {
	prepared, payload, err := c.prepareRequest(req)
	if err != nil {
		return nil, err
	}
	prepared.Headers, err = c.SignHeaders(prepared, payload)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 1; attempt <= c.cfg.MaxRetries+1; attempt++ {
		resp, err := c.send(ctx, prepared, payload)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !shouldRetry(err) || attempt > c.cfg.MaxRetries {
			break
		}
		backoff := time.Duration(1<<(attempt-1)) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	return nil, lastErr
}

func (c *Client) send(ctx context.Context, req APIRequest, payload string) (resp *Response, err error) {
	baseURL := c.cfg.OverrideServiceURL
	if baseURL == "" {
		host, _ := endpointHost(c.cfg.Region)
		baseURL = "https://" + host
	}
	uri := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(req.Path, "/")
	if len(req.QueryParams) > 0 {
		uri += "?" + encodeQuery(req.QueryParams)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, uri, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := httpResp.Body.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode >= 400 {
		return nil, &HTTPError{StatusCode: httpResp.StatusCode, Header: httpResp.Header.Clone(), Body: body}
	}
	return &Response{StatusCode: httpResp.StatusCode, Header: httpResp.Header.Clone(), Body: body}, nil
}

func (c *Client) prepareRequest(req APIRequest) (APIRequest, string, error) {
	headers := make(map[string]string, len(req.Headers))
	for k, v := range req.Headers {
		headers[k] = v
	}
	prepared := req
	prepared.Headers = headers
	prepared.Path = c.resolvePath(req.Path)

	payload := ""
	if req.Payload != nil {
		switch p := req.Payload.(type) {
		case string:
			payload = p
		case []byte:
			payload = string(p)
		default:
			b, err := json.Marshal(req.Payload)
			if err != nil {
				return APIRequest{}, "", fmt.Errorf("marshal payload: %w", err)
			}
			payload = string(b)
		}
	}
	return prepared, payload, nil
}

func (c *Client) resolvePath(fragment string) string {
	clean := strings.TrimLeft(fragment, "/")
	if strings.HasPrefix(clean, "sandbox/") || strings.HasPrefix(clean, "live/") || strings.HasPrefix(clean, apiVersion+"/") {
		return clean
	}
	if isEnvSpecificPublicKeyID(c.cfg.PublicKeyID) {
		return apiVersion + "/" + clean
	}
	env := "live"
	if c.cfg.Environment == EnvironmentSandbox {
		env = "sandbox"
	}
	return env + "/" + apiVersion + "/" + clean
}

func encodeQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(parts, "&")
}

func shouldRetry(err error) bool {
	httpErr := new(HTTPError)
	if !AsHTTPError(err, &httpErr) {
		return false
	}
	return httpErr.StatusCode == http.StatusRequestTimeout ||
		httpErr.StatusCode == http.StatusTooEarly ||
		httpErr.StatusCode == http.StatusTooManyRequests ||
		httpErr.StatusCode >= 500
}

func parseArrayQuery(values []string) string {
	return strings.Join(values, ",")
}

func (c *Client) GetSignedHeaders(req APIRequest) (map[string]string, error) {
	prepared, payload, err := c.prepareRequest(req)
	if err != nil {
		return nil, err
	}
	return c.SignHeaders(prepared, payload)
}

func (c *Client) GetAuthorizationToken(ctx context.Context, mwsAuthToken, merchantID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "authorizationTokens/" + mwsAuthToken, Headers: headers, QueryParams: map[string]string{"merchantId": merchantID}})
}

func (c *Client) DeliveryTrackers(ctx context.Context, payload DeliveryTrackersRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "deliveryTrackers", Payload: payload, Headers: headers})
}

func (c *Client) MerchantScan(ctx context.Context, payload MerchantScanRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "in-store/merchantScan", Payload: payload, Headers: headers})
}

func (c *Client) InStoreCharge(ctx context.Context, payload InStoreChargeRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "in-store/charge", Payload: payload, Headers: headers})
}

func (c *Client) InStoreRefund(ctx context.Context, payload InStoreRefundRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "in-store/refund", Payload: payload, Headers: headers})
}

func (c *Client) GetBuyer(ctx context.Context, buyerToken string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "buyers/" + buyerToken, Headers: headers})
}

func (c *Client) CreateCheckoutSession(ctx context.Context, payload CreateCheckoutSessionRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "checkoutSessions", Payload: payload, Headers: headers})
}

func (c *Client) GetCheckoutSession(ctx context.Context, checkoutSessionID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "checkoutSessions/" + checkoutSessionID, Headers: headers})
}

func (c *Client) UpdateCheckoutSession(ctx context.Context, checkoutSessionID string, payload UpdateCheckoutSessionRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPatch, Path: "checkoutSessions/" + checkoutSessionID, Payload: payload, Headers: headers})
}

func (c *Client) CompleteCheckoutSession(ctx context.Context, checkoutSessionID string, payload CompleteCheckoutSessionRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "checkoutSessions/" + checkoutSessionID + "/complete", Payload: payload, Headers: headers})
}

func (c *Client) FinalizeCheckoutSession(ctx context.Context, checkoutSessionID string, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "checkoutSessions/" + checkoutSessionID + "/finalize", Payload: payload, Headers: headers})
}

func (c *Client) GetChargePermission(ctx context.Context, chargePermissionID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "chargePermissions/" + chargePermissionID, Headers: headers})
}

func (c *Client) UpdateChargePermission(ctx context.Context, chargePermissionID string, payload UpdateChargePermissionRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPatch, Path: "chargePermissions/" + chargePermissionID, Payload: payload, Headers: headers})
}

func (c *Client) CloseChargePermission(ctx context.Context, chargePermissionID string, payload CloseChargePermissionRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodDelete, Path: "chargePermissions/" + chargePermissionID + "/close", Payload: payload, Headers: headers})
}

func (c *Client) CreateCharge(ctx context.Context, payload CreateChargeRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "charges", Payload: payload, Headers: headers})
}

func (c *Client) GetCharge(ctx context.Context, chargeID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "charges/" + chargeID, Headers: headers})
}

func (c *Client) UpdateCharge(ctx context.Context, chargeID string, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPatch, Path: "charges/" + chargeID, Payload: payload, Headers: headers})
}

func (c *Client) CaptureCharge(ctx context.Context, chargeID string, payload CaptureChargeRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "charges/" + chargeID + "/capture", Payload: payload, Headers: headers})
}

func (c *Client) CancelCharge(ctx context.Context, chargeID string, payload CancelChargeRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodDelete, Path: "charges/" + chargeID + "/cancel", Payload: payload, Headers: headers})
}

func (c *Client) CreateRefund(ctx context.Context, payload CreateRefundRequest, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "refunds", Payload: payload, Headers: headers})
}

func (c *Client) GetRefund(ctx context.Context, refundID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "refunds/" + refundID, Headers: headers})
}

func (c *Client) GetReports(ctx context.Context, queryParams map[string]string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "reports", QueryParams: queryParams, Headers: headers})
}

func (c *Client) GetReportByID(ctx context.Context, reportID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "reports/" + reportID, Headers: headers})
}

func (c *Client) GetReportDocument(ctx context.Context, reportDocumentID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "report-documents/" + reportDocumentID, Headers: headers})
}

func (c *Client) GetReportSchedules(ctx context.Context, reportTypes []string, headers map[string]string) (*Response, error) {
	qp := map[string]string{}
	if len(reportTypes) > 0 {
		qp["reportTypes"] = parseArrayQuery(reportTypes)
	}
	if len(qp) == 0 {
		qp = nil
	}
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "report-schedules", QueryParams: qp, Headers: headers})
}

func (c *Client) GetReportScheduleByID(ctx context.Context, reportScheduleID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "report-schedules/" + reportScheduleID, Headers: headers})
}

func (c *Client) CreateReport(ctx context.Context, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "reports", Payload: payload, Headers: headers})
}

func (c *Client) CreateReportSchedule(ctx context.Context, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "report-schedules", Payload: payload, Headers: headers})
}

func (c *Client) CancelReportSchedule(ctx context.Context, reportScheduleID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodDelete, Path: "report-schedules/" + reportScheduleID, Headers: headers})
}

func (c *Client) CreateMerchantAccount(ctx context.Context, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "merchantAccounts", Payload: payload, Headers: headers})
}

func (c *Client) UpdateMerchantAccount(ctx context.Context, merchantAccountID string, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPatch, Path: "merchantAccounts/" + merchantAccountID, Payload: payload, Headers: headers})
}

func (c *Client) DeleteMerchantAccount(ctx context.Context, merchantAccountID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodDelete, Path: "merchantAccounts/" + merchantAccountID, Headers: headers})
}

func (c *Client) MerchantAccountClaim(ctx context.Context, merchantAccountID string, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "merchantAccounts/" + merchantAccountID + "/claim", Payload: payload, Headers: headers})
}

func (c *Client) CreateDispute(ctx context.Context, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "disputes", Payload: payload, Headers: headers})
}

func (c *Client) GetDispute(ctx context.Context, disputeID string, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodGet, Path: "disputes/" + disputeID, Headers: headers})
}

func (c *Client) UpdateDispute(ctx context.Context, disputeID string, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPatch, Path: "disputes/" + disputeID, Payload: payload, Headers: headers})
}

func (c *Client) ContestDispute(ctx context.Context, disputeID string, payload any, headers map[string]string) (*Response, error) {
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "disputes/" + disputeID + "/contest", Payload: payload, Headers: headers})
}

func (c *Client) UploadFile(ctx context.Context, payload any, headers map[string]string) (*Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}
	return c.APICall(ctx, APIRequest{Method: http.MethodPost, Path: "files", Payload: strings.TrimSpace(buf.String()), Headers: headers})
}

func IntToString(v int) string { return strconv.Itoa(v) }
