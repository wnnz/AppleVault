package appstore

import (
	"context"
	"fmt"
	"io"

	"AppleVault/src/backend/internal/ipatool/sap/assets"
	"AppleVault/src/backend/internal/ipatool/sap/machine"
)

type storeAgentDecrypter struct {
	agent *machine.StoreAgent
}

func defaultMacPackageDecrypterFactory(ctx context.Context, hardwareID, dpInfo []byte) (macPackageDecrypter, error) {
	bundle, err := assets.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load SAP assets: %w", err)
	}

	agent, err := machine.OpenStoreAgent(ctx, bundle, hardwareID, dpInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open StoreAgent: %w", err)
	}

	return &storeAgentDecrypter{agent: agent}, nil
}

func (d *storeAgentDecrypter) Decrypt(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	written, err := d.agent.Decrypt(ctx, dst, src)
	if err != nil {
		return written, fmt.Errorf("failed to decrypt with StoreAgent: %w", err)
	}

	return written, nil
}

func (d *storeAgentDecrypter) Close() error {
	if err := d.agent.Close(); err != nil {
		return fmt.Errorf("failed to close StoreAgent: %w", err)
	}

	return nil
}
