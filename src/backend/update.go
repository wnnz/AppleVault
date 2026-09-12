package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	AppVersion        = "1.3.0"
	AppDisplayVersion = "1.3"
	latestReleaseAPI  = "https://api.github.com/repos/wnnz/AppleVault/releases/latest"
	maxReleaseBody    = 1 << 20
)

func (a *App) GetAppVersion() string { return AppDisplayVersion }

func (a *App) CheckForUpdates() (UpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return UpdateInfo{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "AppleVault/"+AppVersion)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	a.settingsMu.RLock()
	proxyEnabled := a.settings.EnableProxy
	proxyAddress := strings.TrimSpace(a.settings.ProxyUrl)
	a.settingsMu.RUnlock()
	if proxyEnabled && proxyAddress != "" {
		proxyURL, parseErr := url.Parse(proxyAddress)
		if parseErr != nil {
			return UpdateInfo{}, fmt.Errorf("代理地址格式无效: %w", parseErr)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	} else {
		transport.Proxy = nil
	}
	response, err := (&http.Client{Transport: transport, Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("检查更新失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return UpdateInfo{}, fmt.Errorf("检查更新失败: GitHub 返回 HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxReleaseBody+1))
	if err != nil {
		return UpdateInfo{}, err
	}
	if len(body) > maxReleaseBody {
		return UpdateInfo{}, fmt.Errorf("更新信息响应过大")
	}
	var release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		Body        string `json:"body"`
		HTMLURL     string `json:"html_url"`
		PublishedAt string `json:"published_at"`
	}
	if err = json.Unmarshal(body, &release); err != nil {
		return UpdateInfo{}, fmt.Errorf("解析更新信息失败: %w", err)
	}
	latest := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	if _, err = parseVersion(latest); err != nil {
		return UpdateInfo{}, fmt.Errorf("发布版本格式无效: %w", err)
	}
	return UpdateInfo{
		CurrentVersion: AppVersion, LatestVersion: latest,
		Available: isVersionNewer(latest, AppVersion), ReleaseName: release.Name,
		ReleaseNotes: release.Body, ReleaseURL: release.HTMLURL, PublishedAt: release.PublishedAt,
	}, nil
}

func parseVersion(value string) ([3]int, error) {
	var result [3]int
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	value = strings.SplitN(value, "-", 2)[0]
	parts := strings.Split(value, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return result, fmt.Errorf("%q", value)
	}
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return result, fmt.Errorf("%q", value)
		}
		result[index] = number
	}
	return result, nil
}

func isVersionNewer(candidate, current string) bool {
	next, nextErr := parseVersion(candidate)
	now, nowErr := parseVersion(current)
	if nextErr != nil || nowErr != nil {
		return false
	}
	for index := range next {
		if next[index] != now[index] {
			return next[index] > now[index]
		}
	}
	return false
}
