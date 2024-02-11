package traefik_plugin_rewrite_response_headers

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

// Rewrite holds one rewrite header configuration.
type Rewrite struct {
	Header      string `json:"header,omitempty"`
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

// Config holds the plugin configuration.
type Config struct {
	Rewrites []Rewrite `json:"rewrites,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewrite struct {
	header      string
	regex       *regexp.Regexp
	replacement string
}

type rewriteHeader struct {
	name     string
	next     http.Handler
	rewrites []rewrite
}

// New creates and returns a new plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rewrites := make([]rewrite, len(config.Rewrites))
	for i, r := range config.Rewrites {
		regex, err := regexp.Compile(r.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", r.Regex, err)
		}
		rewrites[i] = rewrite{header: r.Header, regex: regex, replacement: r.Replacement}
	}

	return &rewriteHeader{
		name:     name,
		next:     next,
		rewrites: rewrites,
	}, nil
}

func (r *rewriteHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rrw := &responseRewriter{writer: rw, rewrites: r.rewrites, requestHost: req.Host}
	r.next.ServeHTTP(rrw, req)
}

type responseRewriter struct {
	writer      http.ResponseWriter
	rewrites    []rewrite
	requestHost string
}

func (r *responseRewriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseRewriter) Write(p []byte) (int, error) {
	return r.writer.Write(p)
}

func (r *responseRewriter) WriteHeader(statusCode int) {
	for _, rewrite := range r.rewrites {
		headers := r.writer.Header().Values(rewrite.header)

		if len(headers) == 0 {
			continue
		}

		r.writer.Header().Del(rewrite.header)

		for _, header := range headers {
			value := rewrite.regex.ReplaceAllString(header, rewrite.replacement)

			if strings.Contains(value, "{RequestHost}") {
				value = strings.ReplaceAll(value, "{RequestHost}", r.requestHost)
			}

			r.writer.Header().Add(rewrite.header, value)
		}
	}

	r.writer.WriteHeader(statusCode)
}

func (r *responseRewriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.writer.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer is not a hijacker")
	}

	return hijacker.Hijack()
}

func (r *responseRewriter) Flush() {
	if flusher, ok := r.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
