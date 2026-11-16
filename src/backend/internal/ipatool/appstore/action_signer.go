package appstore

import (
	"context"
	"fmt"

	"AppleVault/src/backend/internal/ipatool/sap"
	"AppleVault/src/backend/internal/ipatool/http"
)

type SAPConfig struct {
	AuthEndpoint   string
	SetupURL       string
	CertificateURL string
	Version        uint32
}

type ActionSigner interface {
	http.ActionSigner
	Close() error
}

type ActionSignerFactory func(config SAPConfig, machineID []byte) (ActionSigner, error)

func frameworkActionSignerFactory(client http.Executor, ctx context.Context) ActionSignerFactory {
	return func(config SAPConfig, machineID []byte) (ActionSigner, error) {
		signer, err := sap.NewSigner(ctx, sap.Config{SetupURL: config.SetupURL, CertificateURL: config.CertificateURL, Version: config.Version, HardwareID: machineID, HTTPClient: client})
		if err != nil {
			return nil, fmt.Errorf("create SAP signer: %w", err)
		}

		return signer, nil
	}
}
