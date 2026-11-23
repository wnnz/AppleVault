package appstore

import "time"

type App struct {
	ID             int64      `json:"trackId,omitempty"`
	BundleID       string     `json:"bundleId,omitempty"`
	Name           string     `json:"trackName,omitempty"`
	Version        string     `json:"version,omitempty"`
	Price          float64    `json:"price,omitempty"`
	FormattedPrice string     `json:"formattedPrice,omitempty"`
	Currency       string     `json:"currency,omitempty"`
	PurchaseDate   time.Time  `json:"purchaseDate,omitzero"`
	Platforms      []Platform `json:"platforms,omitzero"`
}

type VersionHistoryInfo struct {
	App                App
	LatestVersion      string
	VersionIdentifiers []string
}

type VersionDetails struct {
	VersionID     string
	VersionString string
	Success       bool
	Error         string
}

type Apps []App
