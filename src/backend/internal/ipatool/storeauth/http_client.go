//nolint:wsl // Trust setup follows certificate validation order.
package storeauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const appleRootDERHash = "b0b1730ecbc7ff4505142c49f1295e6eda6bcaed7e2c68c5be91b5a11001f024"

//go:embed apple_root_ca.pem
var appleRootCAPEM []byte

// NewHTTPClient returns the isolated client used for Apple authentication.
// base, when non-nil, is cloned as the transport instead of http.DefaultTransport.
func NewHTTPClient(jar http.CookieJar, base *http.Transport) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if base != nil {
		transport = base.Clone()
	}
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}

	return &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func isUnknownAuthority(err error) bool {
	var unknown x509.UnknownAuthorityError

	return errors.As(err, &unknown)
}

func installAppleRoot(ctx context.Context, client *http.Client) error {
	block, _ := pem.Decode(appleRootCAPEM)
	if block == nil {
		return errors.New("decode embedded Apple root PEM")
	}
	if sum := sha256.Sum256(block.Bytes); hex.EncodeToString(sum[:]) != appleRootDERHash {
		return errors.New("embedded Apple root hash mismatch")
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse Apple root: %w", err)
	}
	if !validAppleRoot(certificate, time.Now()) {
		return errors.New("embedded Apple root is invalid")
	}

	roots, _ := x509.SystemCertPool()
	if roots == nil {
		roots = x509.NewCertPool()
	}
	roots.AddCert(certificate)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		return errors.New("apple authentication transport is unsupported")
	}
	updated := transport.Clone()
	updated.TLSClientConfig = transport.TLSClientConfig.Clone()
	updated.TLSClientConfig.RootCAs = roots
	updated.TLSClientConfig.MinVersion = tls.VersionTLS12
	transport.CloseIdleConnections()
	client.Transport = updated

	return nil
}

func validAppleRoot(certificate *x509.Certificate, now time.Time) bool {
	if certificate == nil || !certificate.IsCA || !certificate.BasicConstraintsValid ||
		certificate.KeyUsage&x509.KeyUsageCertSign == 0 || certificate.Subject.CommonName != "Apple Root CA" ||
		!bytes.Equal(certificate.RawSubject, certificate.RawIssuer) || now.Before(certificate.NotBefore) || now.After(certificate.NotAfter) {
		return false
	}

	for _, organization := range certificate.Subject.Organization {
		if organization == "Apple Inc." {
			return true
		}
	}

	return false
}
