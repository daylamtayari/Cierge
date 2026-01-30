package resy

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// This is an API key that Resy requires on all requests,
// authenticated or not, and has no tie to a user
// From what I can tell, this has remained unchanged
// across the last 5 years but could change at any point
//
// Recommendation is to store this value and regularly
// check if it is still valid by making a request to the
// Geoip endpoint as it is the lightest unauth API request
// that requires an ApiKey and if an unauthorized error
// is returned, fetch the API key
const DefaultApiKey = "VbWk7s3L4KiK5fzlO7JD3Q5EYolJI7n5"

var (
	ErrApiKeyNotFound       = errors.New("could not find API key in module file")
	ErrApiKeyEndNotFound    = errors.New("could not find end of API key value in module file")
	ErrAppModuleNotFound    = errors.New("unable to find app module file in HTML")
	ErrModuleSuffixNotFound = errors.New("unable to find filetype suffix for module")
)

type Tokens struct {
	ApiKey string
	Token  string
}

// Retrieves the Resy API key using a provided http client
// If nil is passed, the default http client is used
func FetchApiKey(httpClient *http.Client) (string, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	res, err := httpClient.Get(Host)
	if err != nil {
		return "", err
	}

	defer res.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Extract the app module filename
	// Format: modules/app.[hash].js
	modulePrefix := []byte("modules/app.")
	moduleSuffix := []byte(".js")

	startIndex := bytes.Index(body, modulePrefix)
	if startIndex == -1 {
		return "", ErrAppModuleNotFound
	}
	searchStart := startIndex + len(modulePrefix)
	endIndex := bytes.Index(body[searchStart:], moduleSuffix)
	if endIndex == -1 {
		return "", ErrModuleSuffixNotFound
	}

	moduleFile := string(body[startIndex : searchStart+endIndex+len(moduleSuffix)])

	res, err = httpClient.Get(Host + "/" + moduleFile)
	if err != nil {
		return "", err
	}

	defer res.Body.Close() //nolint:errcheck
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Extract the API key
	apiKeyPrefix := []byte(`apiKey:"`)
	quoteDelim := []byte(`"`)

	keyStartIndex := bytes.Index(body, apiKeyPrefix)
	if keyStartIndex == -1 {
		return "", ErrApiKeyNotFound
	}
	keySearchStart := keyStartIndex + len(apiKeyPrefix)
	keyEndIndex := bytes.Index(body[keySearchStart:], quoteDelim)
	if keyEndIndex == -1 {
		return "", ErrApiKeyEndNotFound
	}

	apiKey := string(body[keySearchStart : keySearchStart+keyEndIndex])
	return apiKey, nil
}
