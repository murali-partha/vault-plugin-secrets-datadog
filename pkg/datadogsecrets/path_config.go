package datadogsecrets

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configStoragePath = "config"
)

type datadogConfig struct {
	ApiKey         string `json:"api_key"`
	ApplicationKey string `json:"app_key"`
	Host           string `json:"host"`
}

func pathConfig(b *datadogBackend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"datadog_api_key": {
				Type:        framework.TypeString,
				Description: "The API Key to access Datadog API",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Datadog API Key",
					Sensitive: true,
				},
			},
			"datadog_app_key": {
				Type:        framework.TypeString,
				Description: "The Application Key to access Datadog API",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Datadog Application key secret",
					Sensitive: true,
				},
			},
			"host": {
				Type:        framework.TypeString,
				Description: "The host for calling Datadog API",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "host",
					Sensitive: false,
				},
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},
	}
}

func (b *datadogBackend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		b.Logger().Error("error reading configuration", "error", err)
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"datadog_api_key": config.ApiKey,
			"datadog_app_key": config.ApplicationKey,
			"host":            config.Host,
		},
	}, nil
}

func (b *datadogBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		b.Logger().Error("error reading configuration", "error", err)
		return nil, err
	}

	createOperation := (req.Operation == logical.CreateOperation)

	if config == nil {
		if !createOperation {
			return nil, errors.New("config not found during update operation")
		}
		config = new(datadogConfig)
	}

	if keyId, ok := data.GetOk("datadog_api_key"); ok {
		config.ApiKey = keyId.(string)
	} else if !ok && createOperation {
		return nil, fmt.Errorf("missing datadog_api_key in configuration")
	}
	if secret, ok := data.GetOk("datadog_app_key"); ok {
		config.ApplicationKey = secret.(string)
	} else if !ok && createOperation {
		return nil, fmt.Errorf("missing datadog_app_key in configuration")
	}

	if url, ok := data.GetOk("host"); ok {
		config.Host = url.(string)
	} else if !ok && createOperation {
		config.Host = data.GetDefaultOrZero("host").(string)
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	b.reset()

	return nil, nil
}

func (b *datadogBackend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, configStoragePath)

	if err == nil {
		b.reset()
	}

	return nil, err
}

func getConfig(ctx context.Context, s logical.Storage) (*datadogConfig, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}

	config := new(datadogConfig)

	if entry != nil {
		if err := entry.DecodeJSON(&config); err != nil {
			return nil, fmt.Errorf("error reading root configuration: %w", err)
		}
	}

	return config, nil
}
