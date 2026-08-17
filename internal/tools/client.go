package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// default kmesh address
const (
	DefaultStatusAddr = "localhost:15200"
	requestTimeout    = 10 * time.Second
)

// baseURL = kmesh url
type StatusClient struct {
	baseURL    string
	httpClient *http.Client
}

// initialise ,set baseurl and return client object
func NewStatusClient(baseURL string) *StatusClient {
	if baseURL == "" {
		baseURL = DefaultStatusAddr
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	return &StatusClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

// GET request common to all tools
func (c *StatusClient) Get(ctx context.Context, path string) ([]byte, int, error) {
	//create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("building request for %s: %w", path, err)
	}

	//send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("sending request to %s: %w", path, err)
	}

	//close resource associated with the HTTP response
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response from %s: %w", path, err)
	}

	return body, resp.StatusCode, nil
}

type Toolset struct {
	client *StatusClient
}

// initialise and return Toolset object
func NewToolSet(baseURL string) *Toolset {
	return &Toolset{
		client: NewStatusClient(baseURL),
	}
}
