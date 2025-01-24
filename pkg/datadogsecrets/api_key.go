package datadogsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	datadogApiKeyType = "datadog_apikey"
)

type dataDogApiKey struct {
	KeyId  string `json:"key_id"`
	ApiKey string `json:"api_key"`
	Name   string `json:"usage_count"`
}

func (b *datadogBackend) dataDogApiKey() *framework.Secret {
	return &framework.Secret{
		Type: datadogApiKeyType,
		Fields: map[string]*framework.FieldSchema{
			"key_id": {
				Type:        framework.TypeString,
				Description: "Datadog API Key ID",
			},
			"api_key": {
				Type:        framework.TypeString,
				Description: "Datadog API Key",
			},
		},
		Revoke: b.tokenRevoke,
	}
}

func CreateApiKeyCredential(
	ctx context.Context, client *datadogApiKeyClient,
) (*dataDogApiKey, error) {
	keyId, apiKey, err := client.CreateApiKey()
	if err != nil {
		client.log.Error("error creating datadog API Key: %w", err)
		return nil, fmt.Errorf("error creating datadog API Key: %w", err)
	}

	return &dataDogApiKey{
		KeyId:  keyId,
		ApiKey: apiKey,
	}, nil
}

func (b *datadogBackend) tokenRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		client.log.Error("error while getting client: %w", err)
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	keyId := ""
	if keyIdRaw, ok := req.Secret.InternalData["key_id"]; ok {
		keyId, ok = keyIdRaw.(string)
		if !ok {
			client.log.Error("invalid value for token in secret internal data")
			return nil, fmt.Errorf("invalid value for token in secret internal data")
		}
	}

	if _, err := client.DeleteApiKey(keyId); err != nil {
		client.log.Error("error revoking user token: %w", err)
		return nil, fmt.Errorf("error revoking user token: %w", err)
	}

	return nil, nil
}
