package reservation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrUnsuccessfulStatusCode = errors.New("non-200 HTTP code returned")
)

// Notifies the server about the status of a job
func notifyServer(ctx context.Context, serverEndpoint string, callbackSecret string, output []byte) error {
	if !strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint += "/"
	}

	reqUrl := serverEndpoint + "internal/job/status"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(output))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+callbackSecret)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != 200 {
		return fmt.Errorf("%w: %d", ErrUnsuccessfulStatusCode, res.StatusCode)
	}
	return nil
}
