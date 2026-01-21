package resy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Resy API host
const Host = "https://api.resy.com"

// Generic popular user agent to use as default
// if not specified by a user
const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"

var (
	ErrBadRequest      = errors.New("bad or malformed request")
	ErrBadGateway      = errors.New("bad gateway - likely due to malformed input")
	ErrNotFound        = errors.New("not found")
	ErrPaymentRequired = errors.New("payment required")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUnhandledStatus = errors.New("unhandled status code returned")
)

type Tokens struct {
	ApiKey string
	Token  string
}

type Client struct {
	client *http.Client
}

type transport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	for key, value := range t.headers {
		req.Header.Add(key, value)
	}
	return t.base.RoundTrip(req)
}

// Creates a new Resy API client. It accepts an `http.Client` value
// that will be used as the base HTTP client and will have the
// authorization added to. If nil is provided, `http.DefaultClient`
// is used.
// Tokens include the generic Resy API key and the user's token.
// A user agent value to be added to requests is also accepted and if
// an empty string is provided, a popular generic user agent is used.
func NewClient(httpClient *http.Client, tokens Tokens, userAgent string) *Client {
	trans := http.DefaultTransport
	if httpClient == nil {
		httpClient = http.DefaultClient
	} else if t := httpClient.Transport; t != nil {
		trans = httpClient.Transport
	}

	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	httpClient.Transport = &transport{
		base: trans,
		headers: map[string]string{
			"Authorization":         "ResyAPI api_key=\"" + tokens.ApiKey + "\"",
			"X-Resy-Auth-Token":     tokens.Token,
			"X-Resy-Universal-Auth": tokens.Token,
			// User Agent requires as Resy will throw 500s if not included
			"User-Agent": userAgent,
		},
	}

	return &Client{
		client: httpClient,
	}
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

	var body []byte
	if res.ContentLength != 0 && (res.StatusCode == 200 || res.StatusCode == 201 || res.StatusCode == 400) {
		defer res.Body.Close() //nolint: errcheck
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		// Handle 400s differently as opposed to other error status
		// codes as usually 400s return information in the body about
		// what was wrong in the request and this way it allows the error
		// message to contain a wrapped ErrBadRequest that can be unwrapped
		// and identified, as well as the body response that can be used for
		// debugging and understanding the error.
		if res.StatusCode == 400 {
			return fmt.Errorf("%w: %v", ErrBadRequest, body)
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
	case 201:
		return nil
	case 402:
		return ErrPaymentRequired
	case 404:
		return ErrNotFound
	case 419:
		// Resy returns the 419 status code for 'Unauthorized' error messages
		// why not a 401 or 403? don't ask me...
		return ErrUnauthorized
	case 502:
		// Identified that Resy will return 502s on various errors related to
		// malformed or unexpected input values
		return ErrBadGateway
	default:
		return fmt.Errorf("%w: %d", ErrUnhandledStatus, res.StatusCode)
	}
}
