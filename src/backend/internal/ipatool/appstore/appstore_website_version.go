package appstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	gohttp "net/http"
	"net/url"
	"strconv"
	"strings"
)

const maxWebsiteVersionResponseSize = 2 << 20

type websiteVersionLookupResponse struct {
	Data []websiteVersionLookupItem `json:"data"`
}

type websiteVersionLookupItem struct {
	ID         string                         `json:"id"`
	Type       string                         `json:"type"`
	Attributes websiteVersionLookupAttributes `json:"attributes"`
}

type websiteVersionLookupAttributes struct {
	PlatformAttributes struct {
		IOS websiteIOSVersion `json:"ios"`
	} `json:"platformAttributes"`
}

type websiteIOSVersion struct {
	BundleID          string                    `json:"bundleId"`
	ExternalVersionID platformVersionExternalID `json:"externalVersionId"`
}

func (t *appstore) lookupWebsiteExternalVersionID(app App, countryCode string) (string, error) {
	params := url.Values{}
	params.Add("platform", "web")
	params.Add("additionalPlatforms", "iphone,ipad")
	params.Add("l", "en-GB")

	endpoint := fmt.Sprintf(
		"https://apps.apple.com/api/apps/v1/catalog/%s/apps/%d?%s",
		strings.ToLower(countryCode), app.ID, params.Encode(),
	)
	req, err := t.httpClient.NewRequest(gohttp.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create website version request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	res, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("website version request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != gohttp.StatusOK {
		return "", fmt.Errorf("website version lookup failed (HTTP %d)", res.StatusCode)
	}

	limited := io.LimitReader(res.Body, maxWebsiteVersionResponseSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("failed to read website version response: %w", err)
	}
	if len(body) > maxWebsiteVersionResponseSize {
		return "", errors.New("website version response is too large")
	}

	var payload websiteVersionLookupResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("failed to decode website version response: %w", err)
	}
	if len(payload.Data) == 0 {
		return "", errors.New("website version lookup returned no matching app")
	}
	if len(payload.Data) != 1 {
		return "", errors.New("website version lookup returned duplicate app records")
	}

	item := payload.Data[0]
	if item.Type != "apps" {
		return "", errors.New("website version lookup returned an unexpected resource type")
	}
	if item.ID != strconv.FormatInt(app.ID, 10) {
		return "", errors.New("website version lookup returned no matching app")
	}

	version := item.Attributes.PlatformAttributes.IOS
	if version.BundleID == "" || version.BundleID != app.BundleID {
		return "", errors.New("website version lookup returned a missing or mismatched bundle identifier")
	}

	externalVersionID := string(version.ExternalVersionID)
	parsedID, err := strconv.ParseUint(externalVersionID, 10, 64)
	if err != nil || parsedID == 0 || strconv.FormatUint(parsedID, 10) != externalVersionID {
		return "", errors.New("website version lookup returned an invalid external version id")
	}

	return externalVersionID, nil
}
