package datadogsecrets

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
)

type datadogApiKeyClient struct {
	ctx    context.Context
	client *datadog.APIClient

	log hclog.Logger
}

func newApiClient(config *datadogConfig, logger hclog.Logger) (*datadogApiKeyClient, error) {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	context := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: config.ApiKey,
			},
			"appKeyAuth": {
				Key: config.ApplicationKey,
			},
		},
	)

	datadogConfig := datadog.NewConfiguration()
	datadogConfig.Host = config.Host
	apiClient := datadog.NewAPIClient(datadogConfig)

	return &datadogApiKeyClient{ctx: context, client: apiClient, log: logger}, nil
}

func (c *datadogApiKeyClient) CreateApiKey() (string, string, error) {
	usersApi := datadogV2.NewKeyManagementApi(c.client)

	apiKeyName := "vault-api-key" + uuid.New().String()

	body := *datadogV2.NewAPIKeyCreateRequest(*datadogV2.NewAPIKeyCreateData(*datadogV2.NewAPIKeyCreateAttributes(apiKeyName), datadogV2.APIKeysType("api_keys")))

	resp, _, err := usersApi.CreateAPIKey(c.ctx, body)
	if err != nil {
		c.log.Error("error creating datadog API Key", "error", err)
		return "", "", err
	}

	return *resp.Data.Id, *resp.Data.Attributes.Key, nil
}

func (c *datadogApiKeyClient) DeleteApiKey(keyID string) (*http.Response, error) {
	usersApi := datadogV2.NewKeyManagementApi(c.client)

	httpResponse, err := usersApi.DeleteAPIKey(c.ctx, keyID)
	if err != nil {
		c.log.Error("error deleting datadog API Key", "error", err)
		return nil, err
	}

	return httpResponse, nil
}
