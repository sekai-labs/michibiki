package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPClient_HTMLResponseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<!DOCTYPE html><html><head><title>Login</title></head><body>Login required</body></html>"))
	}))
	defer ts.Close()

	c := NewHTTPClient(HTTPOptions{BaseURL: ts.URL})

	var target map[string]any
	err := c.Do(context.Background(), "GET", "/api/interfaces", nil, &target)
	if err == nil {
		t.Fatalf("expected error on HTML response, got nil")
	}

	if !strings.Contains(err.Error(), "HTML") && !strings.Contains(err.Error(), "<!DOCTYPE") {
		t.Fatalf("expected informative error mentioning HTML or snippet, got: %v", err)
	}
	var strTarget string
	if err := c.Do(context.Background(), "GET", "/api/interfaces", nil, &strTarget); err != nil {
		t.Fatalf("expected string target to succeed, got: %v", err)
	}
	if !strings.Contains(strTarget, "Login") {
		t.Fatalf("unexpected string target content: %s", strTarget)
	}
}

func TestHTTPClient_ValidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","count":42}`))
	}))
	defer ts.Close()

	c := NewHTTPClient(HTTPOptions{BaseURL: ts.URL})

	var target struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	err := c.Do(context.Background(), "GET", "/api/status", nil, &target)
	if err != nil {
		t.Fatalf("expected success on valid JSON, got: %v", err)
	}
	if target.Status != "ok" || target.Count != 42 {
		t.Fatalf("unexpected parsed result: %+v", target)
	}
}
func TestHTTPClient_RedirectCredentialStripping(t *testing.T) {
	var thirdPartyReceivedAuth string

	thirdPartyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		thirdPartyReceivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":"third_party"}`))
	}))
	defer thirdPartyServer.Close()

	var sameOriginReceivedAuth string
	var sameOriginServer *httptest.Server

	sameOriginServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect-third-party":
			http.Redirect(w, r, thirdPartyServer.URL+"/target", http.StatusFound)
		case "/redirect-same-origin":
			http.Redirect(w, r, sameOriginServer.URL+"/same-origin-target", http.StatusFound)
		case "/same-origin-target":
			sameOriginReceivedAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"received":"same_origin"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer sameOriginServer.Close()

	t.Run("strips credentials on cross-host redirect", func(t *testing.T) {
		c := NewHTTPClient(HTTPOptions{
			BaseURL: sameOriginServer.URL,
			Token:   "super-secret-token",
		})

		var target map[string]any
		err := c.Do(context.Background(), "GET", "/redirect-third-party", nil, &target)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if thirdPartyReceivedAuth != "" {
			t.Fatalf("expected Authorization header to be stripped for third-party redirect, but received: %q", thirdPartyReceivedAuth)
		}
	})

	t.Run("preserves credentials on same-origin redirect", func(t *testing.T) {
		c := NewHTTPClient(HTTPOptions{
			BaseURL: sameOriginServer.URL,
			Token:   "super-secret-token",
		})

		var target map[string]any
		err := c.Do(context.Background(), "GET", "/redirect-same-origin", nil, &target)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if sameOriginReceivedAuth != "Bearer super-secret-token" {
			t.Fatalf("expected Authorization header to be preserved on same-origin redirect, got: %q", sameOriginReceivedAuth)
		}
	})
}
