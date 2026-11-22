package appstore

import (
	"context"
	"errors"
	"fmt"
	gohttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"AppleVault/src/backend/internal/ipatool/http"
)

const urlBagCacheTTL = 15 * time.Minute

type cachedURLBag struct {
	bag       urlBag
	expiresAt time.Time
}

type BagInput struct{}

type BagOutput struct {
	// AuthEndpoint is retained for callers that inspect the bag directly.
	AuthEndpoint string
	SAPConfig    SAPConfig
}

func (t *appstore) Bag(input BagInput) (BagOutput, error) {
	macAddr, err := t.machine.MacAddress()
	if err != nil {
		return BagOutput{}, fmt.Errorf("failed to get mac address: %w", err)
	}

	guid, _, err := machineIdentity(macAddr)
	if err != nil {
		return BagOutput{}, err
	}

	return t.bag(nil, guid)
}

func (t *appstore) bag(ctx context.Context, guid string) (BagOutput, error) {
	bag, err := t.fetchURLBag(ctx, guid)
	if err != nil {
		return BagOutput{}, err
	}

	version, err := strconv.ParseUint(bag.SAPVersion, 10, 32)
	if err != nil {
		return BagOutput{}, fmt.Errorf("invalid SAP version %q in bag: %w", bag.SAPVersion, err)
	}

	config := SAPConfig{
		AuthEndpoint:   bag.AuthEndpoint,
		SetupURL:       bag.SAPSetupEndpoint,
		CertificateURL: bag.SAPSetupCertEndpoint,
		Version:        uint32(version),
	}
	if err := validateSAPConfig(config); err != nil {
		return BagOutput{}, err
	}

	authURL, err := url.Parse(config.AuthEndpoint)
	if err != nil {
		return BagOutput{}, fmt.Errorf("failed to parse authentication endpoint: %w", err)
	}

	// The legacy bare path can return unusable responses, including a 301
	// without Location. The native endpoint must retain its distinct path.
	if strings.HasSuffix(strings.ToLower(authURL.Hostname()), "buy.itunes.apple.com") {
		authURL.Path = PrivateAppStoreAPIPathAuth + "/"
	}

	config.AuthEndpoint = authURL.String()

	return BagOutput{AuthEndpoint: config.AuthEndpoint, SAPConfig: config}, nil
}

// Download endpoints do not require the authentication-specific SAP configuration.
func (t *appstore) fetchURLBag(ctx context.Context, guid string) (urlBag, error) {
	if ctx == nil {
		ctx = t.requestContext()
	}

	if err := ctx.Err(); err != nil {
		return urlBag{}, err
	}

	t.bagCacheMu.Lock()
	if cached, ok := t.bagCache[guid]; ok && time.Now().Before(cached.expiresAt) {
		t.bagCacheMu.Unlock()

		return cached.bag, nil
	}
	t.bagCacheMu.Unlock()

	res, err := t.bagClient.Send(t.bagRequest(ctx, guid))
	if err != nil {
		return urlBag{}, fmt.Errorf("failed to send http request: %w", err)
	}

	if res.StatusCode != gohttp.StatusOK {
		return urlBag{}, fmt.Errorf("received unexpected status code: %d", res.StatusCode)
	}

	resolved := res.Data.URLBag
	if resolved.AuthEndpoint == "" && resolved.EntDownloadEndpoint == "" &&
		resolved.RedownloadEndpoint == "" && resolved.UpdateEndpoint == "" &&
		resolved.SAPSetupEndpoint == "" && resolved.SAPSetupCertEndpoint == "" &&
		resolved.SAPVersion == "" {
		resolved = urlBag{
			AuthEndpoint:         res.Data.AuthEndpoint,
			EntDownloadEndpoint:  res.Data.EntDownloadEndpoint,
			RedownloadEndpoint:   res.Data.RedownloadEndpoint,
			UpdateEndpoint:       res.Data.UpdateEndpoint,
			SAPSetupEndpoint:     res.Data.SAPSetupEndpoint,
			SAPSetupCertEndpoint: res.Data.SAPSetupCertEndpoint,
			SAPVersion:           res.Data.SAPVersion,
		}
	}

	t.bagCacheMu.Lock()
	if t.bagCache == nil {
		t.bagCache = make(map[string]cachedURLBag)
	}

	t.bagCache[guid] = cachedURLBag{
		bag:       resolved,
		expiresAt: time.Now().Add(urlBagCacheTTL),
	}
	t.bagCacheMu.Unlock()

	return resolved, nil
}

type bagResult struct {
	URLBag               urlBag `plist:"urlBag,omitempty"`
	AuthEndpoint         string `plist:"authenticateAccount,omitempty"`
	EntDownloadEndpoint  string `plist:"volumeStoreDownloadProduct,omitempty"`
	RedownloadEndpoint   string `plist:"redownloadProduct,omitempty"`
	UpdateEndpoint       string `plist:"updateProduct,omitempty"`
	SAPSetupEndpoint     string `plist:"sign-sap-setup,omitempty"`
	SAPSetupCertEndpoint string `plist:"sign-sap-setup-cert,omitempty"`
	SAPVersion           string `plist:"sign-sap-version,omitempty"`
}

type urlBag struct {
	AuthEndpoint         string `plist:"authenticateAccount,omitempty"`
	EntDownloadEndpoint  string `plist:"volumeStoreDownloadProduct,omitempty"`
	RedownloadEndpoint   string `plist:"redownloadProduct,omitempty"`
	UpdateEndpoint       string `plist:"updateProduct,omitempty"`
	SAPSetupEndpoint     string `plist:"sign-sap-setup,omitempty"`
	SAPSetupCertEndpoint string `plist:"sign-sap-setup-cert,omitempty"`
	SAPVersion           string `plist:"sign-sap-version,omitempty"`
}

func validateSAPConfig(config SAPConfig) error {
	if err := validateAuthenticationEndpoint(config.AuthEndpoint); err != nil {
		return err
	}

	endpoints := []struct {
		name string
		url  string
	}{
		{name: "SAP setup", url: config.SetupURL},
		{name: "SAP setup certificate", url: config.CertificateURL},
	}

	for _, endpoint := range endpoints {
		parsed, err := url.ParseRequestURI(endpoint.url)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("invalid %s endpoint %q in bag", endpoint.name, endpoint.url)
		}
	}

	if config.Version != 200 {
		return fmt.Errorf("unsupported SAP version %d in bag", config.Version)
	}

	return nil
}

func validateAuthenticationEndpoint(endpoint string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.Fragment != "" || parsed.RawPath != "" || (parsed.Port() != "" && parsed.Port() != "443") {
		return errors.New("invalid authentication endpoint")
	}

	host := strings.ToLower(parsed.Hostname())
	if host == PrivateAppStoreNativeDomain {
		switch parsed.Path {
		case PrivateAppStoreNativePath, PrivateAppStoreNativePath + "/", PrivateAppStoreNativePath + "/fast", PrivateAppStoreNativePath + "/fast/":
			return nil
		default:
			return errors.New("unsupported authentication endpoint")
		}
	}

	if host != PrivateAppStoreAPIDomain && !strings.HasSuffix(host, "-buy.itunes.apple.com") {
		return errors.New("unsupported authentication endpoint")
	}

	if parsed.Path != PrivateAppStoreAPIPathAuth && parsed.Path != PrivateAppStoreAPIPathAuth+"/" {
		return errors.New("unsupported authentication endpoint")
	}

	return nil
}

func (*appstore) bagRequest(ctx context.Context, guid string) http.Request {
	return http.Request{
		Context:        ctx,
		URL:            fmt.Sprintf("https://%s%s?ix=6&guid=%s", PrivateInitDomain, PrivateInitPath, guid),
		Method:         http.MethodGET,
		ResponseFormat: http.ResponseFormatXML,
		Headers: map[string]string{
			"Accept":     "application/xml",
			"User-Agent": storeAuthenticationUserAgent,
		},
	}
}
