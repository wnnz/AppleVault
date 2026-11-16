# ipatool-derived App Store module

This directory contains the runtime modules needed by AppleVault, derived from
`majd/ipatool` v2.8.1. The Cobra CLI (`cmd/`), logging (`pkg/log`), the
`main.go` entry point, the `internal/store/profilegen` generator, tests, and
generated mocks are intentionally excluded. Module layout:

- `pkg/appstore` -> `appstore`
- `pkg/http` -> `http`
- `pkg/keychain` -> `keychain`
- `pkg/util/...` -> `util/...`
- `internal/sap/...` -> `sap/...`
- `internal/store/...` -> `store/...`
- `internal/storeauth` -> `storeauth`

AppleVault-specific changes:

- `appstore.Args.Context` is the default context for all requests made by the
  store instance; a non-nil `Context` on an input struct still wins. It is
  used wherever upstream falls back to `context.Background()` (Store backend
  executor, native login, SAP signer, download paths).
- `appstore.Args.Transport *http.Transport` is cloned as the base transport
  for `store.New` and `storeauth.NewHTTPClient` instead of
  `http.DefaultTransport`, so the app-level proxy settings apply. SAP asset
  and IPSW downloads keep their own clients and are not proxied.
- The Apple Root CA is embedded in `storeauth/apple_root_ca.pem` (DER SHA-256
  `B0B1730E...01F024`) instead of being downloaded, which fails from some
  networks.
- `storefront_builtin.go` provides a built-in storefront table;
  `discoverAvailableStorefronts` falls back to it whenever the live
  `availableStoreFronts` request fails and merges missing entries on success.
  `CountryCodeFromStoreFront` remains exported for account region display.
- `GetVersionMetadataOutput.FileSize int64` reports the remote IPA size read
  through the partial-zip range reader.
- Download integrity checks: the temporary file's MD5 is verified against the
  checksum from the download response (hex or base64), and the patched IPA is
  validated as a ZIP containing a top-level `Payload/*.app/Info.plist`;
  invalid files are removed. Resume/Content-Range handling is upstream's
  `downloadResponseRange` (HTTP 200 restart, 416 completion, range
  validation).
- `appstore/app.go` drops the zerolog `MarshalZerolog*` methods and import.

Copyright (c) 2021 majd

Licensed under the MIT License. The full license text is distributed at
`third_party/ipatool-LICENSE.txt`.
