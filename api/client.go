package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrBadRequest      = errors.New("bad or malformed request")
	ErrNotFound        = errors.New("not found")
	ErrServerError     = errors.New("server encountered internal error")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUnhandledStatus = errors.New("unhandled status code returned")
	ErrInvalidHost     = errors.New("invalid host value provided")
)

type Client struct {
	client *http.Client
	host   string
}

type transport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	return t.base.RoundTrip(req)
}

// Creates a new Cierge API client. It accepts an `http.Client` value that will
// be used as the base HTTP client and will have the authentication added to. If
// nil is provided, `http.DefaultClient` is used.
// The host value represents the hostname and optionally scheme, of the Cierge
// instance to interact with. If no scheme is provided, HTTPS will be assumed
// The API key is the user's API key to use to create an authenticated client.
// If no API key is provided, the client will not be authenticated.
func NewClient(httpClient *http.Client, host string, apiKey string) (*Client, error) {
	if !strings.Contains(host, "https://") && !strings.Contains(host, "http://") {
		host = "https://" + host
	}
	hostUrl, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	hostUrl.Path = ""
	hostUrl.RawQuery = ""
	hostUrl.Fragment = ""
	if hostUrl.Host == "" {
		return nil, ErrInvalidHost
	}
	clientHost := strings.TrimRight(hostUrl.String(), "/") // Remove trailing slash

	trans := http.DefaultTransport
	if httpClient == nil {
		httpClient = http.DefaultClient
	} else if t := httpClient.Transport; t != nil {
		trans = httpClient.Transport
	}

	headers := map[string]string{}
	if apiKey != "" {
		headers["Authorization"] = "api " + apiKey
	}

	httpClient.Transport = &transport{
		base:    trans,
		headers: headers,
	}

	return &Client{
		client: httpClient,
		host:   clientHost,
	}, nil
}

// Wraps the http.NewRequest function to marshal a provided JSON value and set it
// as the request body. Also sets the content type to JSON.
func (c *Client) NewJsonRequest(method string, url string, jsonValue any) (*http.Request, error) {
	reqBody, err := json.Marshal(jsonValue)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// Do performs an API request, handles the response,
// and unmarshals the response into a given interface.
// The value to unmarshal must be a pointer to an interface.
// If a pointer to a byte array is provided, the returned value
// will be the value of the body.
func (c *Client) Do(req *http.Request, v any) error {
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint: errcheck

	var body []byte
	if res.ContentLength != 0 && res.StatusCode == 200 {
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if _, ok := v.(*[]byte); ok {
			// If a byte array is provided, the body value
			// is returned directly and not unmarshalled
			*v.(*[]byte) = body
		} else if v != nil {
			err = json.Unmarshal(body, &v)
		}
		if err != nil {
			return err
		}
	}

	switch res.StatusCode {
	case 200:
		return nil
	case 401:
		return ErrUnauthenticated
	case 403:
		return ErrUnauthorized
	case 404:
		return ErrNotFound
	case 500:
		return ErrServerError
	default:
		return fmt.Errorf("%w: %d", ErrUnhandledStatus, res.StatusCode)
	}
}
