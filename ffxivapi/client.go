package ffxivapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	chttp "github.com/c032/go-http"
)

const httpUserAgent = "github.com/c032/ffxiv-world-status/discord"

func request[T any](ac *apiClient, method string, path string, body []byte) (*T, error) {
	var (
		err error

		req *http.Request
	)
	req, err = http.NewRequest(http.MethodGet, ac.worldsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	if ac.token != "" {
		req.Header.Set("X-Api-Key", ac.token)
	}

	var resp *http.Response

	resp, err = ac.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var result T

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("could not decode JSON response: %w", err)
	}

	return &result, nil
}

type ClientOptions struct {
	BaseURL string
	Token   string
}

func NewClient(options ClientOptions) (Client, error) {
	c, err := chttp.NewClient(httpUserAgent)
	if err != nil {
		return nil, fmt.Errorf("could not create API client: %w", err)
	}

	ac := &apiClient{
		c:          c,
		token:      strings.TrimSpace(options.Token),
		rawBaseURL: options.BaseURL,
	}

	err = ac.init()
	if err != nil {
		return nil, fmt.Errorf("could nit initialize API client: %w", err)
	}

	return ac, nil
}

type Client interface {
	Worlds() (*WorldsResponse, error)
}

var _ Client = (*apiClient)(nil)

type apiClient struct {
	c chttp.Client

	token string

	rawBaseURL string
	baseURL    *url.URL

	worldsURL *url.URL
}

func (ac *apiClient) init() error {
	var err error

	if ac.baseURL == nil {
		var baseURL *url.URL

		baseURL, err = url.Parse(ac.rawBaseURL)
		if err != nil {
			return fmt.Errorf("could not parse base URL: %w", err)
		}

		if !strings.HasSuffix(baseURL.Path, "/") {
			return fmt.Errorf("expected base URL's path to end with a `/`")
		}

		ac.baseURL = baseURL
	}

	ac.worldsURL, err = ac.resolve("worlds")
	if err != nil {
		return fmt.Errorf("could not initialize worlds URL: %w", err)
	}

	return nil
}

func (ac *apiClient) resolve(urlStr string) (*url.URL, error) {
	var (
		err       error
		parsedURL *url.URL
	)

	parsedURL, err = ac.baseURL.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("could not resolve URL (%#v): %w", urlStr, err)
	}

	if !sameOrigin(*ac.baseURL, *parsedURL) {
		return nil, fmt.Errorf("resolved URL does not have the same origin")
	}

	return parsedURL, nil
}

func (ac *apiClient) Worlds() (*WorldsResponse, error) {
	worldsResponse, err := request[WorldsResponse](ac, http.MethodGet, ac.worldsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch worlds: %w", err)
	}

	return worldsResponse, nil
}
