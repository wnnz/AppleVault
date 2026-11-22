package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	versionMetadataCacheFile     = "version_metadata_cache.json"
	maxVersionMetadataCacheItems = 5000
)

type cachedVersionMetadata struct {
	ExternalVersionID string `json:"externalVersionID"`
	DisplayVersion    string `json:"displayVersion"`
	ReleaseDate       string `json:"releaseDate"`
	FileSize          int64  `json:"fileSize"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type versionMetadataCachePayload struct {
	Entries map[string]cachedVersionMetadata `json:"entries"`
}

func versionMetadataIDKey(appID int64, versionID string) string {
	return fmt.Sprintf("id:%d:%s", appID, strings.TrimSpace(versionID))
}

func versionMetadataBundleKey(bundleID, versionID string) string {
	return fmt.Sprintf("bundle:%s:%s", strings.ToLower(strings.TrimSpace(bundleID)), strings.TrimSpace(versionID))
}

func versionMetadataResultFromCache(versionID string, entry cachedVersionMetadata) VersionMetadataResult {
	releaseDate := entry.ReleaseDate
	if releaseDate == "" {
		releaseDate = "-"
	}
	return VersionMetadataResult{
		ExternalVersionID: versionID,
		DisplayVersion:    entry.DisplayVersion,
		ReleaseDate:       releaseDate,
		FileSize:          entry.FileSize,
		DisplayFileSize:   formatBytes(entry.FileSize),
		Success:           true,
	}
}

func (a *App) ensureVersionMetadataCacheLoadedLocked() {
	dataDir := a.getDataDir()
	if a.versionMetaLoaded && a.versionMetaDataDir == dataDir {
		return
	}

	a.versionMetaEntries = make(map[string]cachedVersionMetadata)
	a.versionMetaLoaded = true
	a.versionMetaDataDir = dataDir

	path := filepath.Join(dataDir, versionMetadataCacheFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var payload versionMetadataCachePayload
	if err := json.Unmarshal(data, &payload); err == nil && payload.Entries != nil {
		a.versionMetaEntries = payload.Entries
	}
}

func (a *App) lookupVersionMetadataLocked(bundleID string, appID int64, versionID string) (VersionMetadataResult, bool) {
	trimmedVersionID := strings.TrimSpace(versionID)
	if trimmedVersionID == "" {
		return VersionMetadataResult{}, false
	}

	if appID > 0 {
		if entry, ok := a.versionMetaEntries[versionMetadataIDKey(appID, trimmedVersionID)]; ok && entry.DisplayVersion != "" {
			return versionMetadataResultFromCache(trimmedVersionID, entry), true
		}
	}

	if strings.TrimSpace(bundleID) != "" {
		if entry, ok := a.versionMetaEntries[versionMetadataBundleKey(bundleID, trimmedVersionID)]; ok && entry.DisplayVersion != "" {
			return versionMetadataResultFromCache(trimmedVersionID, entry), true
		}
	}

	return VersionMetadataResult{}, false
}

func (a *App) getCachedVersionMetadata(bundleID string, appID int64, versionID string) (VersionMetadataResult, bool) {
	if strings.TrimSpace(versionID) == "" {
		return VersionMetadataResult{}, false
	}

	a.versionMetaMu.Lock()
	defer a.versionMetaMu.Unlock()

	a.ensureVersionMetadataCacheLoadedLocked()
	return a.lookupVersionMetadataLocked(bundleID, appID, versionID)
}

func (a *App) getCachedVersionMetadataBatch(bundleID string, appID int64, versionIDs []string) map[string]VersionMetadataResult {
	if len(versionIDs) == 0 {
		return nil
	}

	a.versionMetaMu.Lock()
	defer a.versionMetaMu.Unlock()

	a.ensureVersionMetadataCacheLoadedLocked()
	var results map[string]VersionMetadataResult
	for _, versionID := range versionIDs {
		res, ok := a.lookupVersionMetadataLocked(bundleID, appID, versionID)
		if !ok {
			continue
		}
		if results == nil {
			results = make(map[string]VersionMetadataResult)
		}
		results[versionID] = res
	}
	return results
}

func (a *App) putCachedVersionMetadata(bundleID string, appID int64, result VersionMetadataResult) {
	trimmedVersionID := strings.TrimSpace(result.ExternalVersionID)
	trimmedBundleID := strings.TrimSpace(bundleID)
	if !result.Success || trimmedVersionID == "" || strings.TrimSpace(result.DisplayVersion) == "" {
		return
	}
	if appID <= 0 && trimmedBundleID == "" {
		return
	}

	a.versionMetaMu.Lock()
	defer a.versionMetaMu.Unlock()

	a.ensureVersionMetadataCacheLoadedLocked()
	entry := cachedVersionMetadata{
		ExternalVersionID: trimmedVersionID,
		DisplayVersion:    result.DisplayVersion,
		ReleaseDate:       result.ReleaseDate,
		FileSize:          result.FileSize,
		UpdatedAt:         time.Now().UnixNano(),
	}

	if appID > 0 {
		a.versionMetaEntries[versionMetadataIDKey(appID, trimmedVersionID)] = entry
	}
	if trimmedBundleID != "" {
		a.versionMetaEntries[versionMetadataBundleKey(trimmedBundleID, trimmedVersionID)] = entry
	}

	if len(a.versionMetaEntries) > maxVersionMetadataCacheItems {
		type keyAge struct {
			key       string
			updatedAt int64
		}
		items := make([]keyAge, 0, len(a.versionMetaEntries))
		for k, v := range a.versionMetaEntries {
			items = append(items, keyAge{key: k, updatedAt: v.UpdatedAt})
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].updatedAt < items[j].updatedAt
		})
		toDelete := len(a.versionMetaEntries) - maxVersionMetadataCacheItems
		for i := 0; i < toDelete; i++ {
			delete(a.versionMetaEntries, items[i].key)
		}
	}

	payload := versionMetadataCachePayload{Entries: a.versionMetaEntries}
	if data, err := json.Marshal(payload); err == nil {
		_ = writeFileAtomic(filepath.Join(a.versionMetaDataDir, versionMetadataCacheFile), data, 0600)
	}
}
