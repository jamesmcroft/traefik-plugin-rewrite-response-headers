package traefik_plugin_rewrite_response_headers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	target := "https://127.0.0.1"

	testCases := []struct {
		description             string
		rewrites                []Rewrite
		responseHeaders         http.Header
		expectedResponseHeaders http.Header
	}{
		{
			description: "should replace http by https in Operation-Location header",
			rewrites: []Rewrite{
				{
					Header:      "Operation-Location",
					Regex:       "^http://(.+)$",
					Replacement: "https://$1",
				},
			},
			responseHeaders: map[string][]string{
				"Operation-Location": {"http://example.com/page?query=1"},
			},
			expectedResponseHeaders: map[string][]string{
				"Operation-Location": {"https://example.com/page?query=1"},
			},
		},
		{
			description: "should replace {RequestHost} token with the request host in header",
			rewrites: []Rewrite{
				{
					Header:      "Location",
					Regex:       "^http://(.+?)/(.+)$",
					Replacement: "https://{RequestHost}/$2",
				},
			},
			responseHeaders: map[string][]string{
				"Location": {"http://example.com/page?query=1"},
			},
			expectedResponseHeaders: map[string][]string{
				"Location": {target + "/page?query=1"},
			},
		},
		{
			description: "should not replace if the header does not exist",
			rewrites: []Rewrite{
				{
					Header:      "Operation-Location",
					Regex:       "^http://(.+)$",
					Replacement: "https://$1",
				},
			},
			responseHeaders:         map[string][]string{},
			expectedResponseHeaders: map[string][]string{},
		},
		{
			description: "should replace multiple headers",
			rewrites: []Rewrite{
				{
					Header:      "Operation-Location",
					Regex:       "^http://(.+)$",
					Replacement: "https://$1",
				},
				{
					Header:      "Location",
					Regex:       "^http://(.+)$",
					Replacement: "https://$1",
				},
			},
			responseHeaders: map[string][]string{
				"Operation-Location": {"http://example.com/page?query=1"},
				"Location":           {"http://example.com"},
			},
			expectedResponseHeaders: map[string][]string{
				"Operation-Location": {"https://example.com/page?query=1"},
				"Location":           {"https://example.com"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			config := &Config{Rewrites: testCase.rewrites}

			next := func(responseWriter http.ResponseWriter, request *http.Request) {
				for headerKey, headerValues := range testCase.responseHeaders {
					for _, headerValue := range headerValues {
						responseWriter.Header().Add(headerKey, headerValue)
					}
				}
				responseWriter.WriteHeader(http.StatusOK)
			}

			rewriteHeader, err := New(context.Background(), http.HandlerFunc(next), config, "rewriteHeader")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, target, nil)

			rewriteHeader.ServeHTTP(recorder, request)
			for headerKey, expectedHeaderValues := range testCase.expectedResponseHeaders {
				actualHeaderValues := recorder.Header().Values(headerKey)
				if !testEquals(actualHeaderValues, expectedHeaderValues) {
					t.Errorf("expected %s to be %v, got %v", headerKey, expectedHeaderValues, actualHeaderValues)
				}
			}
		})
	}
}

func testEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
