package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rivflux/rivian/pkg/types"
)

// Client represents a Rivian HTTP client
type Client struct {
	httpClient *http.Client
	baseURL    string
	debug      bool
}

// NewClient creates a new HTTP client
func NewClient(baseURL string, debug bool) *Client {
	return &Client{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		debug:      debug,
	}
}

// DoRequest performs an HTTP request and parses the JSON response
func (c *Client) DoRequest(method, path, body string, headers map[string]string, result interface{}) error {
	req, err := http.NewRequest(method, c.baseURL+path, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set default headers
	for k, v := range types.DefaultHeaders {
		req.Header.Set(k, v)
	}

	// Set additional headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	if c.debug {
		fmt.Printf("Response: %s\n", string(respBody))
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	return nil
}

// GetCSRFToken gets a CSRF token from the Rivian API
func (c *Client) GetCSRFToken() (*types.CSRFResponse, error) {
	csrfQuery := `{
		"operationName": "CreateCSRFToken",
		"variables": null,
		"query": "mutation CreateCSRFToken { createCsrfToken { __typename csrfToken appSessionToken } }"
	}`

	var csrfResp types.CSRFResponse
	if err := c.DoRequest("POST", "/graphql", csrfQuery, nil, &csrfResp); err != nil {
		return nil, err
	}

	return &csrfResp, nil
} 