// Package store provides the portable, framework-backed App Store request engine.
package store

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	apphttp "AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/sap/assets"
	"AppleVault/src/backend/internal/ipatool/store/emulator"
	"AppleVault/src/backend/internal/ipatool/store/profile"
)

type Backend struct {
	jar           http.CookieJar
	macAddress    func() (string, error)
	mu            sync.Mutex
	image         []byte
	bag           map[string]any
	transport     *http.Transport
	authTransport *http.Transport
	ctx           context.Context
}

type executor struct {
	backend   *Backend
	operation string
}

// New creates the Store backend. ctx is the default context for requests that
// do not carry one; base, when non-nil, is cloned as the transport instead of
// http.DefaultTransport.
func New(jar http.CookieJar, macAddress func() (string, error), ctx context.Context, base *http.Transport) *Backend {
	if ctx == nil {
		ctx = context.Background()
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if base != nil {
		transport = base.Clone()
	}
	transport.ResponseHeaderTimeout = 30 * time.Second
	transport.IdleConnTimeout = 90 * time.Second
	authTransport := transport.Clone()
	authTransport.DisableKeepAlives = true

	return &Backend{jar: jar, macAddress: macAddress, transport: transport, authTransport: authTransport, ctx: ctx}
}

// Client shares session cookies and bag data while isolating each guest VM.
func (b *Backend) Client(operation string) apphttp.Executor {
	return &executor{backend: b, operation: operation}
}

func (e *executor) Send(request apphttp.Request, body []byte) (*http.Response, error) {
	ctx := request.Context
	if ctx == nil {
		ctx = e.backend.ctx
	}

	if e.operation == "bag-direct" {
		direct, err := http.NewRequestWithContext(ctx, request.Method, request.URL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create Store request: %w", err)
		}

		for key, value := range request.Headers {
			direct.Header.Set(key, value)
		}

		if direct.Header.Get("User-Agent") == "" {
			direct.Header.Set("User-Agent", apphttp.DefaultUserAgent)
		}

		return e.backend.requestTransport(emulator.Request{Operation: "bag"})(direct)
	}

	input := emulator.Request{URL: request.URL, Method: request.Method, Operation: e.operation, Format: string(request.ResponseFormat), Timeout: request.Timeout, Body: body, Headers: make(http.Header), NoRedirects: request.NoRedirects, Raw: request.ResponseFormat == apphttp.ResponseFormatRaw}
	for key, value := range request.Headers {
		input.Headers.Set(key, value)
	}

	if payload, ok := request.Payload.(*apphttp.XMLPayload); ok {
		input.Parameters = payload.Content
		if input.Parameters == nil {
			input.Parameters = map[string]any{}
		}
	}

	if request.ActionSigner != nil {
		input.Sign = request.ActionSigner.Sign
	}

	return e.backend.execute(ctx, input)
}

func (e *executor) Do(request *http.Request) (*http.Response, error) {
	if request == nil {
		return nil, errors.New("store request is nil")
	}

	if e.operation == "sap" {
		return e.backend.requestTransport(emulator.Request{Operation: "authenticate"})(request)
	}

	if e.operation == "transfer" {
		return e.backend.transfer(request)
	}

	body, err := readStoreRequestBody(request)
	if err != nil {
		return nil, err
	}

	ctx := e.backend.resolveContext(request)

	return e.backend.execute(ctx, emulator.Request{URL: request.URL.String(), Method: request.Method, Headers: request.Header.Clone(), Body: body, Operation: e.operation})
}

func readStoreRequestBody(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(io.LimitReader(request.Body, (1<<20)+1))
	_ = request.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("read Store request: %w", err)
	}

	if len(body) > 1<<20 {
		return nil, errors.New("store request body exceeds limit")
	}

	return body, nil
}

func (b *Backend) resolveContext(request *http.Request) context.Context {
	ctx := request.Context()
	if ctx == context.Background() && b.ctx != nil {
		return b.ctx
	}

	return ctx
}

func (b *Backend) hardware(parameters map[string]any) ([]byte, error) {
	if value, ok := parameters["guid"].(string); ok {
		data, err := hex.DecodeString(value)
		if err != nil || len(data) != 6 {
			return nil, errors.New("invalid Store hardware identity")
		}

		return data, nil
	}

	if b.macAddress == nil {
		return nil, errors.New("store machine identity is not configured")
	}

	value, err := b.macAddress()
	if err != nil {
		return nil, fmt.Errorf("get Store machine identity: %w", err)
	}

	data, err := net.ParseMAC(value)
	if err != nil || len(data) != 6 {
		return nil, errors.New("invalid Store machine address")
	}

	return data, nil
}

func (b *Backend) prepare(ctx context.Context) ([]byte, profile.Data, error) {
	prepared, err := profile.Load(ctx)
	if err != nil {
		return nil, profile.Data{}, fmt.Errorf("load Store runtime: %w", err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.image == nil {
		bundle, err := assets.Load(ctx)
		if err != nil {
			return nil, profile.Data{}, fmt.Errorf("load Store runtime: %w", err)
		}

		b.image = bundle.AppleMediaServices
	}

	return b.image, prepared, nil
}

func (b *Backend) execute(ctx context.Context, input emulator.Request) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("store request: %w", err)
	}

	hardware, err := b.hardware(input.Parameters)
	if err != nil {
		return nil, fmt.Errorf("store request: %w", err)
	}

	input.HardwareID = hardware
	if input.Headers == nil {
		input.Headers = make(http.Header)
	}

	if input.Headers.Get("User-Agent") == "" {
		input.Headers.Set("User-Agent", apphttp.DefaultUserAgent)

		if input.Operation == "purchase" {
			input.Headers.Set("User-Agent", "Configurator/2.18 (Macintosh; OS X 15.3.2; 24D81) AppleWebKit/0620.2.4.11.6")
		}
	}

	b.mu.Lock()
	bag := b.bag
	b.mu.Unlock()

	if input.Operation == "purchase" && bag == nil {
		response, err := b.execute(ctx, emulator.Request{URL: "https://init.itunes.apple.com/bag.xml?guid=" + strings.ToUpper(hex.EncodeToString(hardware)), Method: http.MethodGet, Operation: "bag"})
		if err != nil {
			return nil, fmt.Errorf("load purchase bag: %w", err)
		}

		_ = response.Body.Close()

		b.mu.Lock()
		bag = b.bag
		b.mu.Unlock()
	}

	input.Bag = bag

	image, prepared, err := b.prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare Store framework: %w", err)
	}

	guest, err := emulator.Open(ctx, image, prepared)
	if err != nil {
		return nil, fmt.Errorf("open Store framework: %w", err)
	}

	defer guest.Close()

	input.Transport = b.requestTransport(input)

	response, err := guest.Execute(input)
	if err != nil {
		return nil, fmt.Errorf("store request: %w", err)
	}

	if input.Stream {
		return response, nil
	}

	if strings.EqualFold(response.Header.Get("Content-Encoding"), "gzip") {
		reader, e := gzip.NewReader(response.Body)
		if e != nil {
			_ = response.Body.Close()

			return nil, errors.New("invalid compressed Store response")
		}

		data, e := io.ReadAll(io.LimitReader(reader, (32<<20)+1))
		_ = reader.Close()
		_ = response.Body.Close()

		if e != nil {
			return nil, fmt.Errorf("read compressed Store response: %w", e)
		}

		if len(data) > 32<<20 {
			return nil, errors.New("store response exceeds size limit")
		}

		response.Body = io.NopCloser(bytes.NewReader(data))
		response.Header.Del("Content-Encoding")
		response.Header.Del("Content-Length")
		response.ContentLength = int64(len(data))
	}

	if strings.Contains(input.URL, "init.itunes.apple.com/bag.xml") && response.StatusCode == http.StatusOK {
		data, e := io.ReadAll(response.Body)
		_ = response.Body.Close()

		if e != nil {
			return nil, fmt.Errorf("read Store response: %w", e)
		}

		response.Body = io.NopCloser(bytes.NewReader(data))

		var values map[string]any
		if e = apphttp.DecodePropertyList(data, &values); e == nil {
			b.mu.Lock()
			b.bag = values
			b.mu.Unlock()
		}
	}

	return response, nil
}

// CloseIdleConnections releases pooled sockets without interrupting active streams.
func (b *Backend) CloseIdleConnections() {
	b.transport.CloseIdleConnections()
	b.authTransport.CloseIdleConnections()
}

// Each request owns its timeout; compatible requests share only the connection pool.
// Login keeps its existing fresh-connection policy, including framework redirects.
func (b *Backend) requestTransport(input emulator.Request) func(*http.Request) (*http.Response, error) {
	timeout := input.Timeout
	if timeout == 0 && input.Operation == "api" {
		timeout = 30 * time.Second
	}

	transport := b.transport

	if input.Operation == "authenticate" {
		transport = b.authTransport
	}

	client := &http.Client{Transport: transport, Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	return func(request *http.Request) (*http.Response, error) {
		if b.jar != nil && request.Header.Get("Cookie") == "" {
			for _, cookie := range b.jar.Cookies(request.URL) {
				request.AddCookie(cookie)
			}
		}

		response, err := client.Do(request)
		if err == nil && b.jar != nil {
			b.jar.SetCookies(request.URL, response.Cookies())
		}

		if err != nil {
			return nil, fmt.Errorf("store HTTP transport: %w", err)
		}

		return response, nil
	}
}

func allowedTransferURL(value *url.URL) bool {
	if value == nil || value.User != nil || value.Fragment != "" {
		return false
	}

	return value.Hostname() != "" && (value.Scheme == "https" || value.Scheme == "http")
}

func isRedirectStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// transfer executes CDN/artwork streaming and range requests directly on the pooled
// transport without spinning up a Unicorn ARM64 VM per HTTP request.
func (b *Backend) transfer(request *http.Request) (*http.Response, error) {
	ctx := b.resolveContext(request)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("store request: %w", err)
	}

	if !allowedTransferURL(request.URL) {
		return nil, errors.New("unsupported transfer URL")
	}

	body, err := readStoreRequestBody(request)
	if err != nil {
		return nil, err
	}

	send := b.requestTransport(emulator.Request{Operation: "transfer"})
	currentURL := request.URL
	method := request.Method
	if method == "" {
		method = http.MethodGet
	}

	headers := request.Header.Clone()
	if headers == nil {
		headers = make(http.Header)
	}

	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", apphttp.DefaultUserAgent)
	}

	for redirectCount := 0; ; redirectCount++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, currentURL.String(), bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create transfer request: %w", err)
		}

		req.Header = headers.Clone()

		response, err := send(req)
		if err != nil {
			return nil, err
		}

		if !isRedirectStatus(response.StatusCode) {
			return response, nil
		}

		destination, locErr := response.Location()
		if locErr != nil {
			return response, nil
		}

		_ = response.Body.Close()

		if redirectCount >= 4 {
			return nil, errors.New("redirect limit exceeded")
		}

		if !allowedTransferURL(destination) {
			return nil, errors.New("invalid Store redirect destination")
		}

		if !strings.EqualFold(destination.Hostname(), currentURL.Hostname()) {
			for _, sensitive := range []string{"Authorization", "Cookie", "X-Token", "X-Dsid", "Icloud-Dsid"} {
				headers.Del(sensitive)
			}
		}

		if response.StatusCode == http.StatusSeeOther ||
			((response.StatusCode == http.StatusMovedPermanently || response.StatusCode == http.StatusFound) && strings.EqualFold(method, http.MethodPost)) {
			method = http.MethodGet
			body = nil
			headers.Del("Content-Length")
			headers.Del("Content-Type")
		}

		currentURL = destination
	}
}
