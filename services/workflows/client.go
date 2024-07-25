package workflows

import (
	"encoding/json"
	"fmt"
	"kafkatrigger/services/oauth"
	"net/http"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	RunWorkflow(workflowUuid string, payload []byte) (string, error) // returns workflow run UUID
}

type client struct {
	apiClient   *resty.Client
	oauthClient oauth.Client
}

func (c *client) RunWorkflow(workflowUuid string, payload []byte) (string, error) {
	log.Tracef("Running a pipeline, UUID: %s", workflowUuid)
	t, err := c.oauthClient.GetToken()
	if err != nil {
		return "", err
	}

	log.Trace("Succcessfully obtained OAuth token")

	url := fmt.Sprintf("/v1/pipelines/%s/runs/", workflowUuid)
	resp, err := c.apiClient.
		R().
		SetAuthToken(t.AccessToken).
		SetBody(payload).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return "", err
	}

	log.Trace("Succcessfully sent HTTP request")

	if resp.StatusCode() != http.StatusCreated {
		log.Tracef("Response: %s", string(resp.Body()))
		return "", fmt.Errorf("unexpected http status code: %d", resp.StatusCode())
	}

	log.Trace("Pipeline run initiated successfully")

	result := struct {
		WorkflowRunUuid string `json:"pipeline_run_uuid"`
	}{}

	err = json.Unmarshal(resp.Body(), &result)

	return result.WorkflowRunUuid, err
}

func NewClient(baseUrl, clientId, clientSecret, basicAuthUsername, basicAuthPassword string) Client {
	log.Tracef("Creating a new API client, base url: %s", baseUrl)

	apiClient := resty.New().SetBaseURL(baseUrl)
	if len(basicAuthUsername) > 0 || len(basicAuthPassword) > 0 {
		apiClient.SetBasicAuth(basicAuthUsername, basicAuthPassword)
	}

	c := &client{
		apiClient:   apiClient,
		oauthClient: oauth.NewClient(baseUrl, clientId, clientSecret, basicAuthUsername, basicAuthPassword),
	}

	return c
}
