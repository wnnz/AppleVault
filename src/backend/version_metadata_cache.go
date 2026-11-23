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
	versionMetadataCacheFile        = "version_metadata_cache.json"
	maxVersionMetadataCacheItems    = 5000
	defaultVersionMetadataFlushDelay = 800 * time.Millisecond
)

type cachedVersionMetadata struct {
	ExternalVersionID string `json:"externalVersionID"`
	DisplayVersion    string `json:"displayVersion"`
	ReleaseDate       string `json:"releaseDate"`
	FileSize          int64  `json:"fileSize"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type versionMetadataCachePayload struct {
	Entries       map[string]cachedVersionMetadata `json:"entries"`
	BundleToAppID map[string]int64                 `json:"bundleToAppID,omitempty"`
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

	if a.versionMetaFlushTimer != nil {
		a.versionMetaFlushTimer.Stop()
		a.versionMetaFlushTimer = nil
	}
	a.versionMetaEntries = make(map[string]cachedVersionMetadata)
	a.versionMetaBundleToAppID = make(map[string]int64)
	a.versionMetaDirty = false
	a.versionMetaLoaded = true
	a.versionMetaDataDir = dataDir

	path := filepath.Join(dataDir, versionMetadataCacheFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var payload versionMetadataCachePayload
	if unmarshalErr := json.Unmarshal(data, &payload); unmarshalErr == nil {
		if payload.Entries != nil {
			a.versionMetaEntries = payload.Entries
		}
		if payload.BundleToAppID != nil {
			a.versionMetaBundleToAppID = payload.BundleToAppID
		}
		return
	} else {
		backupPath, backupErr := backupCorruptFile(path)
		fields := map[string]interface{}{"backup": backupPath, "error": unmarshalErr.Error()}
		if backupErr != nil {
			fields["backup_error"] = backupErr.Error()
		}
		a.writeDiagnostic("version_metadata_cache.corrupt", fields)
	}
}

func (a *App) lookupVersionMetadataLocked(bundleID string, appID int64, versionID string) (VersionMetadataResult, bool) {
	trimmedVersionID := strings.TrimSpace(versionID)
	if trimmedVersionID == "" {
		return VersionMetadataResult{}, false
	}

	normalizedBundle := strings.ToLower(strings.TrimSpace(bundleID))
	if appID <= 0 && normalizedBundle != "" && a.versionMetaBundleToAppID != nil {
		if mappedID, ok := a.versionMetaBundleToAppID[normalizedBundle]; ok && mappedID > 0 {
			appID = mappedID
		}
	}

	if appID > 0 {
		if entry, ok := a.versionMetaEntries[versionMetadataIDKey(appID, trimmedVersionID)]; ok && entry.DisplayVersion != "" {
			return versionMetadataResultFromCache(trimmedVersionID, entry), true
		}
	}

	if normalizedBundle != "" {
		if entry, ok := a.versionMetaEntries[versionMetadataBundleKey(normalizedBundle, trimmedVersionID)]; ok && entry.DisplayVersion != "" {
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
	normalizedBundle := strings.ToLower(strings.TrimSpace(bundleID))
	if !result.Success || trimmedVersionID == "" || strings.TrimSpace(result.DisplayVersion) == "" {
		return
	}
	if appID <= 0 && normalizedBundle == "" {
		return
	}

	a.versionMetaMu.Lock()
	defer a.versionMetaMu.Unlock()

	a.ensureVersionMetadataCacheLoadedLocked()
	if a.versionMetaBundleToAppID == nil {
		a.versionMetaBundleToAppID = make(map[string]int64)
	}
	if appID <= 0 && normalizedBundle != "" {
		if mappedID, ok := a.versionMetaBundleToAppID[normalizedBundle]; ok && mappedID > 0 {
			appID = mappedID
		}
	}

	entry := cachedVersionMetadata{
		ExternalVersionID: trimmedVersionID,
		DisplayVersion:    result.DisplayVersion,
		ReleaseDate:       result.ReleaseDate,
		FileSize:          result.FileSize,
		UpdatedAt:         time.Now().UnixNano(),
	}

	if appID > 0 {
		a.versionMetaEntries[versionMetadataIDKey(appID, trimmedVersionID)] = entry
		if normalizedBundle != "" {
			a.versionMetaBundleToAppID[normalizedBundle] = appID
			delete(a.versionMetaEntries, versionMetadataBundleKey(normalizedBundle, trimmedVersionID))
		}
	} else if normalizedBundle != "" {
		a.versionMetaEntries[versionMetadataBundleKey(normalizedBundle, trimmedVersionID)] = entry
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

	a.versionMetaDirty = true
	delay := a.versionMetaFlushDelay
	if delay <= 0 {
		delay = defaultVersionMetadataFlushDelay
	}
	if a.versionMetaFlushTimer != nil {
		a.versionMetaFlushTimer.Stop()
	}
	expectedDir := a.versionMetaDataDir
	a.versionMetaFlushTimer = time.AfterFunc(delay, func() {
		a.flushVersionMetadataCacheIfDir(expectedDir)
	})
}

func (a *App) flushVersionMetadataCache() error {
	return a.flushVersionMetadataCacheIfDir("")
}

func (a *App) flushVersionMetadataCacheIfDir(expectedDir string) error {
	a.versionMetaWriteMu.Lock()
	defer a.versionMetaWriteMu.Unlock()

	a.versionMetaMu.Lock()
	if expectedDir != "" && a.versionMetaDataDir != expectedDir {
		a.versionMetaMu.Unlock()
		return nil
	}
	if a.versionMetaFlushTimer != nil {
		a.versionMetaFlushTimer.Stop()
		a.versionMetaFlushTimer = nil
	}
	if !a.versionMetaDirty || a.versionMetaDataDir == "" {
		a.versionMetaMu.Unlock()
		return nil
	}
	entriesSnapshot := make(map[string]cachedVersionMetadata, len(a.versionMetaEntries))
	for k, v := range a.versionMetaEntries {
		entriesSnapshot[k] = v
	}
	bundleSnapshot := make(map[string]int64, len(a.versionMetaBundleToAppID))
	for k, v := range a.versionMetaBundleToAppID {
		bundleSnapshot[k] = v
	}
	targetPath := filepath.Join(a.versionMetaDataDir, versionMetadataCacheFile)
	a.versionMetaDirty = false
	a.versionMetaMu.Unlock()

	payload := versionMetadataCachePayload{
		Entries:       entriesSnapshot,
		BundleToAppID: bundleSnapshot,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		a.versionMetaMu.Lock()
		a.versionMetaDirty = true
		a.versionMetaMu.Unlock()
		return err
	}
	if err := writeFileAtomic(targetPath, data, 0600); err != nil {
		a.versionMetaMu.Lock()
		a.versionMetaDirty = true
		a.versionMetaMu.Unlock()
		return err
	}
	return nil
}
