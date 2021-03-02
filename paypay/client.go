package paypay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
)

// API endpoint base constants
const (
	APIEndpointProduction = "https://api.paypay.ne.jp"
	APIEndpointSandbox    = "https://stg-api.sandbox.paypay.ne.jp"
)

// Client type
type Client struct {
	apiKey         string
	apiSecret      string
	assumeMerchant string
	endpoint       *url.URL
	httpClient     *http.Client
}

// ClientOption type
type ClientOption func(*Client) error

// New returns a new pay client instance.
func New(apiKey, apiSecret, assumeMerchant string, options ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("missing api key")
	}
	if apiSecret == "" {
		return nil, errors.New("missing api secret")
	}
	if assumeMerchant == "" {
		return nil, errors.New("missing merchant")
	}
	c := &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: http.DefaultClient,
	}
	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}
	if c.endpoint == nil {
		u, err := url.Parse(APIEndpointProduction)
		if err != nil {
			return nil, err
		}
		c.endpoint = u
	}
	return c, nil
}

// WithHTTPClient function
func WithHTTPClient(c *http.Client) ClientOption {
	return func(client *Client) error {
		client.httpClient = c
		return nil
	}
}

// WithEndpoint function
func WithEndpoint(endpoint string) ClientOption {
	return func(client *Client) error {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		client.endpoint = u
		return nil
	}
}

// WithSandbox function
func WithSandbox() ClientOption {
	return WithEndpoint(APIEndpointSandbox)
}

// mergeQuery method
func (c *Client) mergeQuery(path string, q interface{}) (string, error) {
	v := reflect.ValueOf(q)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return path, nil
	}

	u, err := url.Parse(path)
	if err != nil {
		return path, err
	}

	qs, err := query.Values(q)
	if err != nil {
		return path, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (c *Client) authHeader(method, path, contentType string, body []byte) string {
	nonce := uuid.New().String()
	epoch := strconv.FormatInt(time.Now().Unix(), 10)

	var (
		authContentType string
		bodyHashPayload string
	)
	switch {
	case len(body) == 0:
		authContentType = "empty"
		bodyHashPayload = "empty"
	default:
		authContentType = contentType
		md5Hash := md5.New()
		md5Hash.Write([]byte(contentType))
		md5Hash.Write(body)
		bodyHashPayload = base64.StdEncoding.EncodeToString(md5Hash.Sum(nil))
	}

	signatureList := strings.Join([]string{
		path,
		method,
		nonce,
		epoch,
		authContentType,
		bodyHashPayload,
	}, "\n")
	hmacHash := hmac.New(sha256.New, []byte(c.apiSecret))
	hmacHash.Write([]byte(signatureList))
	signatureHashPayload := base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))

	return fmt.Sprintf("hmac OPA-Auth:%s", strings.Join([]string{
		c.apiKey,
		signatureHashPayload,
		nonce,
		epoch,
		bodyHashPayload,
	}, ":"))
}

// NewRequest method
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	switch method {
	case http.MethodGet, http.MethodDelete:
		if body != nil {
			merged, err := c.mergeQuery(path, body)
			if err != nil {
				return nil, err
			}
			path = merged
		}
	}
	u, err := c.endpoint.Parse(path)
	if err != nil {
		return nil, err
	}

	var (
		reqBody     io.ReadWriter
		reqBodyData []byte
	)
	switch method {
	case http.MethodPost, http.MethodPut:
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reqBody = bytes.NewBuffer(b)
			reqBodyData = b
		}
	}

	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	if c.assumeMerchant != "" {
		req.Header.Set("X-ASSUME-MERCHANT", c.assumeMerchant)
	}
	const (
		contentType = "application/json"
	)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", c.authHeader(method, strings.Split(path, "?")[0], contentType, reqBodyData))
	return req, nil
}

// Do method
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	defer resp.Body.Close()

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
				return resp, err
			}
		}
	}

	return resp, err
}
