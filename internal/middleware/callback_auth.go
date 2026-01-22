package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	appctx "github.com/daylamtayari/cierge/internal/context"
	"github.com/daylamtayari/cierge/internal/service"
	"github.com/daylamtayari/cierge/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrAlreadyCallbacked = errors.New("callback request already received")
	ErrInvalidAuthHeader = errors.New("authorization header value is invalid")
	ErrInvalidSecret     = errors.New("invalid callback secret")
	ErrNoJobID           = errors.New("no job ID was specified in request")
)

type CallbackAuthMiddleware struct {
	jobService *service.JobService
}

func NewCallbackAuthMiddleware(jobService *service.JobService) *CallbackAuthMiddleware {
	return &CallbackAuthMiddleware{
		jobService: jobService,
	}
}

func (m *CallbackAuthMiddleware) RequireCallbackAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appctx.Logger(c.Request.Context())
		errorCol := appctx.ErrorCollector(c.Request.Context())

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorCol.Add(ErrInvalidAuthHeader, zerolog.InfoLevel, true, nil, "callback request did not contain an authorization header")
			respondUnauthorized(c)
			return
		}

		authToken := strings.TrimPrefix(authHeader, "Bearer ")
		if authToken == authHeader {
			// No Bearer prefix present
			errorCol.Add(ErrInvalidAuthHeader, zerolog.InfoLevel, true, nil, "callback request did not contain a 'Bearer' prefix in the authorization header")
			respondUnauthorized(c)
			return
		}

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("auth_method", string(service.CallbackToken))
		})

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "failed to read the body of a callback request")
			respondUnauthorized(c)
			return
		}
		_ = c.Request.Body.Close()

		// Restore the body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var response struct {
			JobID *uuid.UUID `json:"job_id"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "failed to unmarshal the body of a callback request to retrieve job ID")
			respondUnauthorized(c)
			return
		}
		if response.JobID == nil {
			errorCol.Add(ErrNoJobID, zerolog.InfoLevel, true, nil, "callback request body did not contain job ID field")
			respondUnauthorized(c)
			return
		}

		job, err := m.jobService.GetByID(c, *response.JobID)
		if err != nil && errors.Is(err, service.ErrJobDNE) {
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "job ID specified in callback request does not exist")
			respondUnauthorized(c)
			return
		} else if err != nil {
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve job from job ID for callback request")
			respondInternalServerError(c)
			return
		}

		validToken, err := util.SecureVerifyHash(*job.CallbackSecretHash, authToken)
		if err != nil {
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to verify hash during callback request auth")
			respondInternalServerError(c)
			return
		} else if !validToken {
			errorCol.Add(ErrInvalidSecret, zerolog.InfoLevel, true, nil, "invalid secret provided as token of callback request")
			respondUnauthorized(c)
			return
		}

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("job_id", job.ID.String())
		})

		// If a callback request has already been received
		// for a job, return forbidden
		// A callback request can only be received a single
		// time for a job as any subsequent requests can be
		// assumed to be malicious as the lambda only ever
		// makes a single callback request
		if job.Callbacked {
			errorCol.Add(ErrAlreadyCallbacked, zerolog.InfoLevel, true, nil, "callback request has already been received for this job")
			respondForbidden(c)
			return
		}

		c.Next()
	}
}
