package appstore

import (
	"errors"
	"fmt"
	gohttp "net/http"
	"net/url"
	"strings"

	"AppleVault/src/backend/internal/ipatool/http"
)

type downloadProductEndpoint struct {
	URL                  string
	ExternalVersionIDKey string
}

func (t *appstore) sendDownloadProduct(acc Account, app App, guid, externalVersionID string, platform Platform) (http.Result[downloadResult], error) {
	primary := t.downloadProductRequest(acc, app, guid, externalVersionID, volumeStoreEndpoint(acc))
	res, err := t.downloadClient.Send(primary)
	if err != nil {
		return http.Result[downloadResult]{}, fmt.Errorf("failed to send http request: %w", err)
	}

	if len(res.Data.Items) > 0 || res.Data.FailureType != "" {
		return res, nil
	}

	if platform != "" && platform != PlatformIPhone && platform != PlatformIPad {
		return res, nil
	}

	bag, err := t.bag(guid)
	if err != nil {
		return http.Result[downloadResult]{}, fmt.Errorf("failed to get bag for redownload fallback: %w", err)
	}

	if bag.RedownloadEndpoint == "" {
		return res, nil
	}

	redownload, err := newRedownloadEndpoint(bag.RedownloadEndpoint)
	if err != nil {
		return http.Result[downloadResult]{}, err
	}

	if externalVersionID == "" {
		lookupPlatform := platform
		if lookupPlatform == "" {
			lookupPlatform = PlatformIPhone
		}

		externalVersionID, err = t.lookupLatestExternalVersionID(acc, app, lookupPlatform)
		if err != nil {
			return http.Result[downloadResult]{}, fmt.Errorf("failed to resolve latest iOS version for redownload: %w", err)
		}
	}

	req := t.downloadProductRequest(acc, app, guid, externalVersionID, redownload)
	res, err = t.downloadClient.Send(req)
	if err != nil {
		if isEmptyInternalServerError(err) && bag.UpdateEndpoint != "" {
			return t.sendUpdateProduct(acc, app, guid, externalVersionID, bag.UpdateEndpoint)
		}

		return http.Result[downloadResult]{}, fmt.Errorf("failed to send redownload request: %w", err)
	}

	return res, nil
}

func (t *appstore) sendUpdateProduct(acc Account, app App, guid, externalVersionID, endpoint string) (http.Result[downloadResult], error) {
	update, err := newUpdateEndpoint(endpoint)
	if err != nil {
		return http.Result[downloadResult]{}, err
	}

	req := t.downloadProductRequest(acc, app, guid, externalVersionID, update)
	res, err := t.downloadClient.Send(req)
	if err != nil {
		return http.Result[downloadResult]{}, fmt.Errorf("failed to send update request: %w", err)
	}
	if res.Data.FailureType != "" {
		message := res.Data.CustomerMessage
		if message == "" {
			message = res.Data.FailureType
		}
		return http.Result[downloadResult]{}, fmt.Errorf("received update error: %s", message)
	}
	if res.StatusCode != gohttp.StatusOK {
		return http.Result[downloadResult]{}, fmt.Errorf("received unexpected update status code: %d", res.StatusCode)
	}

	if len(res.Data.Items) != 1 {
		return http.Result[downloadResult]{}, errors.New("update response must contain exactly one item")
	}

	metadata := res.Data.Items[0].Metadata
	if fmt.Sprint(metadata["itemId"]) != fmt.Sprint(app.ID) ||
		fmt.Sprint(metadata["softwareVersionExternalIdentifier"]) != externalVersionID {
		return http.Result[downloadResult]{}, errors.New("update response does not match the requested app or version")
	}
	if fmt.Sprint(metadata["softwareVersionBundleId"]) != app.BundleID {
		return http.Result[downloadResult]{}, errors.New("update response does not match the requested bundle identifier")
	}

	return res, nil
}

func isEmptyInternalServerError(err error) bool {
	var unexpected *http.UnexpectedResponseError
	return errors.As(err, &unexpected) && unexpected.StatusCode == 500 && unexpected.Snippet == ""
}

func volumeStoreEndpoint(acc Account) downloadProductEndpoint {
	podPrefix := ""
	if acc.Pod != "" {
		podPrefix = "p" + acc.Pod + "-"
	}

	return downloadProductEndpoint{
		URL:                  fmt.Sprintf("https://%s%s%s", podPrefix, PrivateAppStoreAPIDomain, PrivateAppStoreAPIPathDownload),
		ExternalVersionIDKey: "externalVersionId",
	}
}

func newRedownloadEndpoint(endpoint string) (downloadProductEndpoint, error) {
	parsed, err := parseDownloadDispatchEndpoint(endpoint)
	if err != nil {
		return downloadProductEndpoint{}, fmt.Errorf("invalid redownload endpoint %q in bag", endpoint)
	}

	if parsed.Path != "/r/redownload" {
		return downloadProductEndpoint{}, fmt.Errorf("unsupported redownload endpoint %q in bag", endpoint)
	}

	return downloadProductEndpoint{URL: endpoint, ExternalVersionIDKey: "appExtVrsId"}, nil
}

func newUpdateEndpoint(endpoint string) (downloadProductEndpoint, error) {
	parsed, err := parseDownloadDispatchEndpoint(endpoint)
	if err != nil || parsed.Path != "/up/updateProduct" {
		return downloadProductEndpoint{}, errors.New("invalid update endpoint in bag")
	}

	return downloadProductEndpoint{URL: endpoint, ExternalVersionIDKey: "appExtVrsId"}, nil
}

func parseDownloadDispatchEndpoint(endpoint string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, errors.New("invalid download dispatch endpoint")
	}
	if strings.ToLower(parsed.Hostname()) != "downloaddispatch.itunes.apple.com" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("invalid download dispatch endpoint")
	}

	return parsed, nil
}

func (*appstore) downloadProductRequest(acc Account, app App, guid, externalVersionID string, endpoint downloadProductEndpoint) http.Request {
	payload := map[string]interface{}{
		"creditDisplay": "",
		"guid":          guid,
		"salableAdamId": app.ID,
		"serialNumber":  "0",
	}

	if externalVersionID != "" {
		payload[endpoint.ExternalVersionIDKey] = externalVersionID
	}

	return http.Request{
		URL:            fmt.Sprintf("%s?guid=%s", endpoint.URL, guid),
		Method:         http.MethodPOST,
		ResponseFormat: http.ResponseFormatXML,
		Headers: map[string]string{
			"Content-Type":        "application/x-apple-plist",
			"iCloud-DSID":         acc.DirectoryServicesID,
			"X-Dsid":              acc.DirectoryServicesID,
			"X-Apple-Store-Front": acc.StoreFront,
			"X-Token":             acc.PasswordToken,
		},
		Payload: &http.XMLPayload{Content: payload},
	}
}
