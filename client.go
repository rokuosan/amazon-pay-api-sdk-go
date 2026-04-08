package amazonpay

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type Client struct {
    cfg        Config
    httpClient *http.Client
    baseURL    string
}

func NewClient(cfg Config) (*Client, error) {
    if err := cfg.validate(); err != nil {
        return nil, err
    }

    return &Client{
        cfg:        cfg,
        httpClient: http.DefaultClient,
        baseURL:    endpoint(cfg.Region, cfg.Environment),
    }, nil
}

type Response struct {
    StatusCode int
    Header     http.Header
    Body       []byte
}

type RequestOption func(*requestOptions)

type requestOptions struct {
    headers map[string]string
}

func WithHeader(k, v string) RequestOption {
    return func(o *requestOptions) {
        if o.headers == nil {
            o.headers = map[string]string{}
        }
        o.headers[k] = v
    }
}

func WithIdempotencyKey(key string) RequestOption {
    return WithHeader("x-amz-pay-idempotency-key", key)
}

func (c *Client) Do(ctx context.Context, method, path string, payload any, opts ...RequestOption) (*Response, error) {
    var body io.Reader
    if payload != nil {
        b, err := json.Marshal(payload)
        if err != nil {
            return nil, err
        }
        body = bytes.NewReader(b)
    }

    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    ro := &requestOptions{}
    for _, opt := range opts {
        opt(ro)
    }

    for k, v := range ro.headers {
        req.Header.Set(k, v)
    }

    if err := c.sign(req, payload); err != nil {
        return nil, fmt.Errorf("sign request: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return &Response{
        StatusCode: resp.StatusCode,
        Header:     resp.Header,
        Body:       respBody,
    }, nil
}

// --- High level APIs ---

func (c *Client) CreateCheckoutSession(ctx context.Context, payload any, opts ...RequestOption) (*Response, error) {
    return c.Do(ctx, http.MethodPost, "/v2/checkoutSessions", payload, opts...)
}

func (c *Client) GetCheckoutSession(ctx context.Context, id string, opts ...RequestOption) (*Response, error) {
    return c.Do(ctx, http.MethodGet, "/v2/checkoutSessions/"+id, nil, opts...)
}

func (c *Client) UpdateCheckoutSession(ctx context.Context, id string, payload any, opts ...RequestOption) (*Response, error) {
    return c.Do(ctx, http.MethodPatch, "/v2/checkoutSessions/"+id, payload, opts...)
}

func (c *Client) CompleteCheckoutSession(ctx context.Context, id string, payload any, opts ...RequestOption) (*Response, error) {
    return c.Do(ctx, http.MethodPost, "/v2/checkoutSessions/"+id+"/complete", payload, opts...)
}

func (c *Client) GetBuyer(ctx context.Context, token string, opts ...RequestOption) (*Response, error) {
    return c.Do(ctx, http.MethodGet, "/v2/buyer/"+token, nil, opts...)
}
