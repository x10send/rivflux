package httpclient

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/x10send/rivflux/pkg/types"
)

const maxResponseBodyBytes = 10 * 1024 * 1024 // 10 MB

// sensitiveHeaders are redacted in debug output to prevent session token leakage.
var sensitiveHeaders = map[string]bool{
	"Authorization": true,
	"Csrf-Token":    true,
	"A-Sess":        true,
	"U-Sess":        true,
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	debug      bool
}

func NewClient(baseURL string, debug bool) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		debug:      debug,
	}
}

func (c *Client) DoRequest(method, path, body string, headers map[string]string, result any) error {
	req, err := http.NewRequest(method, c.baseURL+path, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	for k, v := range types.DefaultHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = io.LimitReader(resp.Body, maxResponseBodyBytes)
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("error creating gzip reader: %v", err)
		}
		defer gz.Close()
		reader = gz
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	if c.debug {
		redacted := make(http.Header)
		for k, v := range req.Header {
			if sensitiveHeaders[k] {
				redacted[k] = []string{"[REDACTED]"}
			} else {
				redacted[k] = v
			}
		}
		fmt.Printf("Request: %s %s\n", req.Method, req.URL)
		fmt.Printf("Request Headers: %v\n", redacted)
		fmt.Printf("Response Status: %d\n", resp.StatusCode)
		fmt.Printf("Response Headers: %v\n", resp.Header)
		fmt.Printf("Response Body: %s\n", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: request failed with status %d", method, c.baseURL+path, resp.StatusCode)
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}
	return nil
}

func (c *Client) GetCSRFToken() (*types.CSRFResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"operationName": "CreateCSRFToken",
		"variables":     nil,
		"query":         "mutation CreateCSRFToken { createCsrfToken { __typename csrfToken appSessionToken } }",
	})

	var csrfResp types.CSRFResponse
	if err := c.DoRequest("POST", "", string(body), nil, &csrfResp); err != nil {
		return nil, err
	}
	return &csrfResp, nil
}
