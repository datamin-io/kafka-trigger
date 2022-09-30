package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	GetToken() (Token, error)
}

type client struct {
	clientId     string
	clientSecret string
	client       *resty.Client
}

func (c *client) GetToken() (Token, error) {
	log.Trace("Requesting a new access token")
	req := c.client.R().
		SetQueryParams(map[string]string{
			"client_id":     c.clientId,
			"client_secret": c.clientSecret,
			"grant_type":    "client_credentials",
		})

	resp, err := req.Get("/v1/oauth/token")

	if err != nil {
		return Token{}, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Tracef("Response: %s", string(resp.Body()))
		return Token{}, fmt.Errorf("unexpected status code while obtaining new access token: %d", resp.StatusCode())
	}

	var result response
	err = json.Unmarshal(resp.Body(), &result)

	if err != nil {
		return Token{}, err
	}

	log.Trace("Successfully obtained a new access token")

	return Token{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn-10) * time.Second),
	}, nil
}

type refreshingClient struct {
	mu           *sync.RWMutex
	currentToken *Token
	innerClient  *client
}

func (c *refreshingClient) GetToken() (Token, error) {
	if c.currentToken == nil || c.currentToken.IsExpired() {
		log.Trace("No current token or expired")
		t, err := c.getNewToken()
		if err != nil {
			return t, err
		}
		c.currentToken = &t
	}

	return *c.currentToken, nil
}

func (c *refreshingClient) getNewToken() (Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t, err := c.innerClient.GetToken()
	return t, err
}

type response struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (t Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func NewClient(baseUrl, clientId, clientSecret, basicAuthUsername, basicAuthPassword string) Client {
	log.Tracef("Creating a new OAuth client, base url: %s", baseUrl)
	restyClient := resty.New().SetBaseURL(baseUrl)
	if len(basicAuthUsername) > 0 || len(basicAuthPassword) > 0 {
		restyClient.SetBasicAuth(basicAuthUsername, basicAuthPassword)
	}
	c := &client{
		client:       restyClient,
		clientId:     clientId,
		clientSecret: clientSecret,
	}

	rc := &refreshingClient{
		mu:          &sync.RWMutex{},
		innerClient: c,
	}

	return rc
}
