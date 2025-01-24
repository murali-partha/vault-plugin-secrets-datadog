package datadogsecrets

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCredentials(b *datadogBackend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the api key",
				Required:    true,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathCredentialsRead,
		},
	}
}

func (b *datadogBackend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("Creating new Datadog API Key")

	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		b.Logger().Error("error getting datadog client", "error", err)
		return nil, err
	}

	apiKey, err := CreateApiKeyCredential(ctx, client)
	if err != nil {
		b.Logger().Error("error creating datadog API Key", "error", err)
		return nil, err
	}

	if apiKey == nil || apiKey.KeyId == "" || apiKey.ApiKey == "" {
		b.Logger().Error("datadog api key is nil", "apiKey: ", apiKey)
		return nil, errors.New("error creating datadog API Key")
	}

	resp := b.Secret(datadogApiKeyType).Response(
		// Data
		map[string]interface{}{
			"key_id": apiKey.KeyId,
			"secret": apiKey.ApiKey,
		},
		// Internal
		map[string]interface{}{
			"key_id": apiKey.KeyId,
		},
	)

	resp.Secret.TTL = 5 * time.Second
	resp.Secret.MaxTTL = 10 * time.Second

	return resp, nil
}
