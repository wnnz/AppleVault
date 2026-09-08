package main

type AccountInfo struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Success bool   `json:"success"`
}

type AppItem struct {
	ID           int64   `json:"id"`
	BundleID     string  `json:"bundleID"`
	Name         string  `json:"name"`
	Version      string  `json:"version"`
	Price        float64 `json:"price"`
	PurchaseDate string  `json:"purchaseDate,omitempty"`
	DisplayPrice string  `json:"displayPrice"`
}

type SearchResult struct {
	Count int       `json:"count"`
	Apps  []AppItem `json:"apps"`
}

type PurchasedResult struct {
	Count      int       `json:"count"`
	TotalCount int       `json:"totalCount"`
	Page       int       `json:"page"`
	Apps       []AppItem `json:"apps"`
}

type VersionsResult struct {
	BundleID                   string   `json:"bundleID"`
	ExternalVersionIdentifiers []string `json:"externalVersionIdentifiers"`
	Success                    bool     `json:"success"`
}

type VersionMetadataResult struct {
	ExternalVersionID string `json:"externalVersionID"`
	DisplayVersion    string `json:"displayVersion"`
	ReleaseDate       string `json:"releaseDate"`
	FileSize          int64  `json:"fileSize"`
	DisplayFileSize   string `json:"displayFileSize"`
	Success           bool   `json:"success"`
}

type DownloadResult struct {
	Output    string `json:"output"`
	Purchased bool   `json:"purchased"`
	Success   bool   `json:"success"`
}

type PurchaseResult struct {
	AlreadyOwned bool `json:"alreadyOwned"`
	Success      bool `json:"success"`
}

type LoginResult struct {
	Success      bool        `json:"success"`
	Requires2FA  bool        `json:"requires2FA"`
	Account      AccountInfo `json:"account"`
	ErrorMessage string      `json:"errorMessage"`
}

type ProxyTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Settings struct {
	KeychainPassphrase string `json:"keychainPassphrase"`
	DefaultDownloadDir string `json:"defaultDownloadDir"`
	DefaultPlatform    string `json:"defaultPlatform"`
	EnableProxy        bool   `json:"enableProxy"`
	ProxyUrl           string `json:"proxyUrl"`
	IpaToolPath        string `json:"ipaToolPath"`
}
