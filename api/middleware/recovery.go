package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"unsafe"

	"github.com/gin-gonic/gin"

	appctx "github.com/daylamtayari/cierge/internal/context"
)

// The Recovery function as well as the brokenPipeError
// function below are modifications or copies of the
// original code from the gin project.
// These were located in the recovery.go file
// https://github.com/gin-gonic/gin/blob/64a6ed9/recovery.go
// These functions are licensed under the MIT license, please
// see the copy of the license below.

// Custom recovery middleware to use the structured
// logger and have greater control.
// Based on Recovery() from gin
// Copyright (c) 2014 Manuel Martínez-Almeida
// Licensed under MIT license (copy located at the bottom of this file)
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger := appctx.Logger(c.Request.Context())

				brokenPipe := brokenPipeError(err)

				if brokenPipe {
					logger.Warn().Err(err.(error)).Msg("broken pipe occurred")
				} else if gin.IsDebugging() {
					// Include request dump if in dev mode
					logger.Error().
						Err(err.(error)).
						Str("path", c.Request.URL.Path).
						Str("request", secureRequestDump(c.Request)).
						Bytes("stack", debug.Stack()).
						Msg("panic recovered")
				} else {
					logger.Error().
						Err(err.(error)).
						Str("path", c.Request.URL.Path).
						Bytes("stack", debug.Stack()).
						Msg("panic recovered")
				}

				if brokenPipe {
					c.Error(err.(error)) //nolint:errcheck
					c.Abort()
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error":      "Internal server error",
						"request_id": appctx.RequestID(c.Request.Context()),
					})
				}
			}
		}()

		c.Next()
	}
}

// Checks if a given error is a broken pipe
// Based on Recovery from gin
// Copyright (c) 2014 Manuel Martínez-Almeida
// Licensed under MIT license (copy located at the bottom of this file)
func brokenPipeError(err any) bool {
	if ne, ok := err.(*net.OpError); ok {
		var se *os.SyscallError
		if errors.As(ne, &se) {
			seStr := strings.ToLower(se.Error())
			if strings.Contains(seStr, "broken pipe") || strings.Contains(seStr, "connection reset by peer") {
				return true
			}
		}
	}
	if e, ok := err.(error); ok && errors.Is(e, http.ErrAbortHandler) {
		return true
	}
	return false
}

// Sanitises an error message
// Based on secureRequestDump from gin
// Copyright (c) 2014 Manuel Martínez-Almeida
// Licensed under MIT license (copy located at the bottom of this file)
func secureRequestDump(r *http.Request) string {
	httpRequest, _ := httputil.DumpRequest(r, false)
	lines := strings.Split(unsafe.String(unsafe.SliceData(httpRequest), len(httpRequest)), "\r\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Authorization:") {
			lines[i] = "Authorization: *****"
		}
	}
	return strings.Join(lines, "\r\n")
}

// MIT license from gin
// https://github.com/gin-gonic/gin/tree/64a6ed9
// The MIT License (MIT)
//
// Copyright (c) 2014 Manuel Martínez-Almeida
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
