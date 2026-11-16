package appstore

import (
	"context"
	gohttp "net/http"
	"sync"
	"time"

	"AppleVault/src/backend/internal/ipatool/store"
	"AppleVault/src/backend/internal/ipatool/storeauth"

	"AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/keychain"
	"AppleVault/src/backend/internal/ipatool/util/machine"
	"AppleVault/src/backend/internal/ipatool/util/operatingsystem"
)

type AppStore interface {
	// Login authenticates with the App Store.
	Login(input LoginInput) (LoginOutput, error)
	// AccountInfo returns the information of the authenticated account.
	AccountInfo() (AccountInfoOutput, error)
	// Revoke revokes the active credentials.
	Revoke() error
	// Lookup looks apps up by bundle identifier or numeric app ID.
	Lookup(input LookupInput) (LookupOutput, error)
	// Search searches the App Store for apps matching the specified term.
	Search(input SearchInput) (SearchOutput, error)
	// OwnedApps lists apps owned by the authenticated account.
	OwnedApps(input OwnedAppsInput) (OwnedAppsOutput, error)
	// Purchase acquires a license for the desired app.
	// Note: only free apps are supported.
	Purchase(input PurchaseInput) error
	// Download downloads the app package from the App Store to the desired location.
	Download(input DownloadInput) (DownloadOutput, error)
	// ReplicateSinf replicates the sinf for the IPA package.
	ReplicateSinf(input ReplicateSinfInput) error
	// VersionHistory lists the available versions of the specified app.
	ListVersions(input ListVersionsInput) (ListVersionsOutput, error)
	// GetVersionMetadata returns the metadata for the specified version.
	GetVersionMetadata(input GetVersionMetadataInput) (GetVersionMetadataOutput, error)
	// Bag fetches the bag which contains endpoint definitions.
	Bag(input BagInput) (BagOutput, error)
}

type appstore struct {
	keychain               keychain.Keychain
	loginClient            http.Client[loginResult]
	searchClient           http.Client[searchResult]
	purchaseClient         http.Client[purchaseResult]
	downloadClient         http.Client[downloadResult]
	kbsyncGenerator        kbsyncGenerator
	kbsyncCache            kbsyncCache
	platformClient         http.Client[platformVersionLookupResult]
	storefrontClient       http.Client[[]byte]
	storefrontsMu          sync.Mutex
	storefrontCountryCodes map[string]string
	bagClient              http.Client[bagResult]
	ownedAppsClient        http.Client[[]byte]
	httpClient             http.Client[interface{}]
	macDecrypterFactory    macPackageDecrypterFactory
	actionSignerFactory    ActionSignerFactory
	authRetrySleep         func(time.Duration)
	machine                machine.Machine
	os                     operatingsystem.OperatingSystem
	authBackend            *store.Backend
	authHTTP               *gohttp.Client
	authStateMu            sync.Mutex
	anisetteState          []byte
	authenticationHeaders  gohttp.Header
	ctx                    context.Context
}

// requestContext returns the store's default context for request work whose
// input did not carry one.
func (t *appstore) requestContext() context.Context {
	if t.ctx != nil {
		return t.ctx
	}

	return context.Background()
}

type Args struct {
	Keychain            keychain.Keychain
	CookieJar           http.CookieJar
	OperatingSystem     operatingsystem.OperatingSystem
	Machine             machine.Machine
	ActionSignerFactory ActionSignerFactory
	// Context is the default context for all requests made by this store
	// instance; a non-nil context on an input struct still wins.
	Context context.Context
	// Transport, when non-nil, is cloned as the base transport for the
	// store's HTTP clients instead of http.DefaultTransport.
	Transport *gohttp.Transport
}

// NewAppStore creates a client using the compiled Store profile.
func NewAppStore(args Args) AppStore {
	var macAddress func() (string, error)
	if args.Machine != nil {
		macAddress = args.Machine.MacAddress
	}

	ctx := args.Context
	if ctx == nil {
		ctx = context.Background()
	}

	backend := store.New(args.CookieJar, macAddress, ctx, args.Transport)
	authHTTP := storeauth.NewHTTPClient(args.CookieJar, args.Transport)
	clientArgs := http.Args{
		CookieJar: args.CookieJar,
		Executor:  backend.Client("request"),
	}

	actionSignerFactory := args.ActionSignerFactory
	if actionSignerFactory == nil {
		actionSignerFactory = frameworkActionSignerFactory(backend.Client("sap"), ctx)
	}

	return &appstore{
		keychain:            args.Keychain,
		loginClient:         http.NewClient[loginResult](http.Args{CookieJar: args.CookieJar, Authentication: true, Executor: backend.Client("authenticate")}),
		searchClient:        http.NewClient[searchResult](clientArgs),
		purchaseClient:      http.NewClient[purchaseResult](http.Args{CookieJar: args.CookieJar, Executor: backend.Client("purchase")}),
		downloadClient:      http.NewClient[downloadResult](clientArgs),
		kbsyncGenerator:     defaultKBSyncGenerator,
		platformClient:      http.NewClient[platformVersionLookupResult](clientArgs),
		storefrontClient:    http.NewClient[[]byte](clientArgs),
		bagClient:           http.NewClient[bagResult](http.Args{CookieJar: args.CookieJar, Timeout: http.DefaultAuthenticationTimeout, Executor: backend.Client("bag-direct")}),
		ownedAppsClient:     http.NewClient[[]byte](clientArgs),
		httpClient:          http.NewClient[interface{}](http.Args{CookieJar: args.CookieJar, Executor: backend.Client("transfer")}),
		macDecrypterFactory: defaultMacPackageDecrypterFactory,
		actionSignerFactory: actionSignerFactory,
		authRetrySleep:      time.Sleep,
		machine:             args.Machine,
		os:                  args.OperatingSystem,
		authBackend:         backend,
		authHTTP:            authHTTP,
		ctx:                 ctx,
	}
}

// NewAppStoreWithSetup is retained for source compatibility. The callback is no
// longer invoked because Store profiles are precomputed.
//
// Deprecated: Use NewAppStore.
func NewAppStoreWithSetup(args Args, _ func()) AppStore { return NewAppStore(args) }
