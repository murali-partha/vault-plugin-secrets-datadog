package datadogsecrets

import (
	"context"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := newBackend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type datadogBackend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *datadogApiKeyClient
}

func newBackend() *datadogBackend {
	b := &datadogBackend{}

	b.Backend = &framework.Backend{
		Paths: framework.PathAppend(
			[]*framework.Path{
				pathConfig(b),
				pathCredentials(b),
			},
		),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{},
			SealWrapStorage: []string{
				"config",
			},
		},
		Secrets: []*framework.Secret{
			b.dataDogApiKey(),
		},
		BackendType: logical.TypeLogical,
	}
	return b
}

func (b *datadogBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

func (b *datadogBackend) getClientCached(ctx context.Context, s logical.Storage) *datadogApiKeyClient {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.client
}

func (b *datadogBackend) getClient(ctx context.Context, s logical.Storage) (*datadogApiKeyClient, error) {
	client := b.getClientCached(ctx, s)
	if client != nil {
		return client, nil
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	if b.client != nil {
		return b.client, nil
	}

	config, err := getConfig(ctx, s)
	if err != nil {
		b.Logger().Error("error getting configuration", "error", err)
		return nil, err
	}

	client, err = newApiClient(config, b.Logger())
	if err != nil {
		b.Logger().Error("error creating new api client from config", "error", err)
		return nil, err
	}

	b.client = client

	return client, nil
}
