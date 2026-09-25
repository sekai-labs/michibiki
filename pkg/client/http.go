package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	baseURL    string
	username   string
	password   string
	apiKey     string
	apiSecret  string
	token      string
	httpClient *http.Client
}

type HTTPOptions struct {
	BaseURL            string
	Username           string
	Password           string
	APIKey             string
	APISecret          string
	Token              string
	InsecureSkipVerify bool
	Timeout            time.Duration
}

func NewHTTPClient(opts HTTPOptions) *HTTPClient {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.InsecureSkipVerify,
		},
	}

	baseURL := strings.TrimRight(opts.BaseURL, "/")
	if baseURL != "" && !strings.HasPrefix(baseURL, "http:") && !strings.HasPrefix(baseURL, "https:") {
		baseURL = "https:" + string([]byte{0x2f, 0x2f}) + baseURL
	}

	clientObj := &HTTPClient{
		baseURL:   baseURL,
		username:  opts.Username,
		password:  opts.Password,
		apiKey:    opts.APIKey,
		apiSecret: opts.APISecret,
		token:     opts.Token,
	}

	clientObj.httpClient = &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			if len(via) > 0 {
				if clientObj.apiKey != "" && clientObj.apiSecret != "" {
					req.SetBasicAuth(clientObj.apiKey, clientObj.apiSecret)
				} else if clientObj.username != "" || clientObj.password != "" {
					req.SetBasicAuth(clientObj.username, clientObj.password)
				} else if clientObj.token != "" {
					req.Header.Set("Authorization", "Bearer "+clientObj.token)
				}
			}
			return nil
		},
	}

	return clientObj
}

func (c *HTTPClient) Do(ctx context.Context, method, path string, body any, target any) error {
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/")

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.apiKey != "" && c.apiSecret != "" {
		req.SetBasicAuth(c.apiKey, c.apiSecret)
	} else if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	if target != nil {
		if strTarget, ok := target.(*string); ok {
			*strTarget = string(respBytes)
			return nil
		}
		if byteTarget, ok := target.(*[]byte); ok {
			*byteTarget = respBytes
			return nil
		}

		contentType := resp.Header.Get("Content-Type")
		trimmed := bytes.TrimSpace(respBytes)
		if strings.HasPrefix(strings.ToLower(contentType), "text/html") ||
			strings.HasPrefix(strings.ToLower(contentType), "text/xml") ||
			(len(trimmed) > 0 && trimmed[0] == '<') {
			snippet := string(trimmed)
			if len(snippet) > 120 {
				snippet = snippet[:120] + "..."
			}
			snippet = strings.ReplaceAll(snippet, "\n", " ")
			snippet = strings.ReplaceAll(snippet, "\r", "")
			lowerTrimmed := strings.ToLower(string(trimmed))
			if strings.Contains(lowerTrimmed, "no-js") || strings.Contains(lowerTrimmed, "login") || strings.Contains(lowerTrimmed, "password") {
				return fmt.Errorf("endpoint returned web GUI login page instead of API response (authentication failed or credentials missing: use -f with API key file or check URL): %s", snippet)
			}
			return fmt.Errorf("endpoint returned HTML/non-JSON response (check URL, credentials, or API path): %s", snippet)
		}
		if err := json.Unmarshal(respBytes, target); err != nil {
			snippet := string(trimmed)
			if len(snippet) > 100 {
				snippet = snippet[:100] + "..."
			}
			snippet = strings.ReplaceAll(snippet, "\n", " ")
			return fmt.Errorf("failed to parse JSON response (%w): %s", err, snippet)
		}
		return nil
	}

	return nil
}
