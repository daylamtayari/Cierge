package resy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

type Client struct {
	client *http.Client
}

type transport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add headers to the request
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	return t.base.RoundTrip(req)
}

// Creates a new Resy API client. It accepts an `http.Client` value
// that will be used as the base HTTP client and will have the
// authentication added to. If nil is provided, `http.DefaultClient`
// is used.
// Tokens include the generic Resy API key and the user's token.
// A user agent value to be added to requests is also accepted and if
// an empty string is provided, a popular generic user agent is used.
// NOTE: A Tokens value that is not scoped to a particular user,
// i.e. only has an ApiKey value but not a Token value, can be used to
// make requests that require the ApiKey but not authentication
// In such cases, the X-Resy-* headers will not be included
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

	headers := map[string]string{
		"Authorization": "ResyAPI api_key=\"" + tokens.ApiKey + "\"",
		// User Agent requires as Resy will throw 500s if not included
		"User-Agent": userAgent,
	}

	if tokens.Token != "" {
		headers["X-Resy-Auth-Token"] = tokens.Token
		headers["X-Resy-Universal-Auth"] = tokens.Token
	}

	httpClient.Transport = &transport{
		base:    trans,
		headers: headers,
	}

	return &Client{
		client: httpClient,
	}
}

// Creates a new http request for a JSON payload
// Marshals the provided jsonValue value, creates
// a new request, and sets the content type to JSON
// Error is only returned if the specified value
// fails to be marshalled or the new request fails to be created
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

// Creates a new http request for a form payload
// Encodes the provided url values into the body,
// creates a new request, and sets the content
// type to url encoded form
func (c *Client) NewFormRequest(method string, url string, form *url.Values) (*http.Request, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

// Wraps DoWithCookies but does not return response cookies,
// only the error that is nil if successful
func (c *Client) Do(req *http.Request, v any) error {
	_, err := c.DoWithCookies(req, v)
	return err
}

// Performs an API request, handles the response,
// and unmarshals the response into a given interface.
// The value to unmarshal must be a pointer to an interface.
// If a pointer to a byte array is provided, the returned value
// will be the value of the body.
// Returns response cookies and an error that is nil if successful
func (c *Client) DoWithCookies(req *http.Request, v any) ([]*http.Cookie, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint: errcheck

	var body []byte
	if res.ContentLength != 0 && (res.StatusCode == 200 || res.StatusCode == 201 || res.StatusCode == 400) {
		// Handle gzip-compressed responses
		reader := res.Body
		if res.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(res.Body)
			if err != nil {
				return nil, err
			}
			defer gzReader.Close() //nolint: errcheck
			reader = gzReader
		}

		body, err = io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		// Handle 400s differently as opposed to other error status
		// codes as usually 400s return information in the body about
		// what was wrong in the request and this way it allows the error
		// message to contain a wrapped ErrBadRequest that can be unwrapped
		// and identified, as well as the body response that can be used for
		// debugging and understanding the error.
		if res.StatusCode == 400 {
			return nil, fmt.Errorf("%w: %v", ErrBadRequest, string(body))
		}

		if _, ok := v.(*[]byte); ok {
			// If a byte array is provided, the body value
			// is returned directly and not unmarshalled
			*v.(*[]byte) = body
		} else if v != nil {
			err = json.Unmarshal(body, &v)
		}
		if err != nil {
			return nil, err
		}
	}

	switch res.StatusCode {
	case 200:
		return res.Cookies(), nil
	case 201:
		return res.Cookies(), nil
	case 402:
		return nil, ErrPaymentRequired
	case 404:
		return nil, ErrNotFound
	case 419:
		// Resy returns the 419 status code for 'Unauthorized' error messages
		// why not a 401 or 403? don't ask me...
		return nil, ErrUnauthorized
	case 502:
		// Identified that Resy will return 502s on various errors related to
		// malformed or unexpected input values
		return nil, ErrBadGateway
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnhandledStatus, res.StatusCode)
	}
}
