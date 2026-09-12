package backend

type AccountInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Region  string `json:"region"`
	Active  bool   `json:"active"`
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

// DeviceInfo 描述一台可供 IPA 安装使用的已连接苹果设备。
type DeviceInfo struct {
	UDID           string `json:"udid"`
	Name           string `json:"name"`
	ProductType    string `json:"productType"`
	ProductVersion string `json:"productVersion"`
	ConnectionType string `json:"connectionType"`
}

// InstallResult 描述 IPA 安装命令的执行结果。
type InstallResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

type DownloadTask struct {
	ID           string `json:"id"`
	AccountID    string `json:"accountID,omitempty"`
	AppName      string `json:"appName"`
	BundleID     string `json:"bundleID"`
	AppID        int64  `json:"appId"`
	Version      string `json:"version"`
	VersionID    string `json:"versionId"`
	FileSize     string `json:"fileSize"`
	TotalBytes   int64  `json:"totalBytes"`
	CurrBytes    int64  `json:"currBytes"`
	Progress     int    `json:"progress"`
	Speed        string `json:"speed"`
	Status       string `json:"status"` // "pending", "downloading", "completed", "error", "canceled"
	OutputPath   string `json:"outputPath"`
	ErrorMessage string `json:"errorMessage"`
	CreatedAt    string `json:"createdAt"`
}

type Settings struct {
	KeychainPassphrase string `json:"keychainPassphrase"`
	DefaultDownloadDir string `json:"defaultDownloadDir"`
	DefaultPlatform    string `json:"defaultPlatform"`
	EnableProxy        bool   `json:"enableProxy"`
	ProxyUrl           string `json:"proxyUrl"`
	// IpaToolPath is kept for settings-file and frontend compatibility.
	IpaToolPath string `json:"ipaToolPath"`
}
