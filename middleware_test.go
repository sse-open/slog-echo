package slogecho

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func mockedMiddleware(t *testing.T, modifiers ...func(Middleware) Middleware) (echo.MiddlewareFunc, *bytes.Buffer) {
	logbuffer := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(logbuffer, nil))

	m := New(logger)

	for _, modifier := range modifiers {
		m = modifier(m)
	}

	return m.EchoMiddleware(), logbuffer
}

func mockGetRequest(t *testing.T, logmiddleware echo.MiddlewareFunc, uri string) {
	req := httptest.NewRequest("GET", uri, nil)
	resp := httptest.NewRecorder()

	err := logmiddleware(func(ctx echo.Context) error {
		return ctx.HTML(http.StatusOK, "<p>example</p>")
	})(echo.New().NewContext(req, resp))
	assert.Nil(t, err)
}

func mockPostRequest(t *testing.T, logmiddleware echo.MiddlewareFunc, uri string, body string) {
	req := httptest.NewRequest("POST", uri, bytes.NewBufferString(body))
	resp := httptest.NewRecorder()

	err := logmiddleware(func(ctx echo.Context) error {
		return ctx.HTML(http.StatusOK, "<p>example</p>")
	})(echo.New().NewContext(req, resp))
	assert.Nil(t, err)
}

func TestMiddleware200(t *testing.T) {
	logmiddleware, logbuffer := mockedMiddleware(t)
	mockPostRequest(t, logmiddleware, "https://example.com/hej/ho?foo=bah", "blahblah")

	assert.Contains(t, logbuffer.String(), "level=INFO msg=REQUEST request.method=POST request.uri=\"https://example.com/hej/ho?foo=bah\" response.status=200")
}

func TestMiddlewareError(t *testing.T) {
	req := httptest.NewRequest("POST", "https://example.com/hej/ho?foo=bah", bytes.NewBufferString("blahblah"))
	resp := httptest.NewRecorder()

	logmiddleware, logbuffer := mockedMiddleware(t)

	err := logmiddleware(func(ctx echo.Context) error {
		return echo.NewHTTPError(500).WithInternal(errors.New("A simulated internal error"))
	})(echo.New().NewContext(req, resp))

	if assert.NotNil(t, err) {
		assert.ErrorContains(t, err, "Internal Server Error")
	}

	assert.Contains(t, logbuffer.String(), "level=ERROR msg=REQUEST request.method=POST request.uri=\"https://example.com/hej/ho?foo=bah\" response.status=500 response.error=\"code=500, message=Internal Server Error, internal=A simulated internal error\"")
}

func TestMiddlewareFilterHealthcheck(t *testing.T) {
	logmiddleware, logbuffer := mockedMiddleware(t, func(m Middleware) Middleware {
		return m.WithFilter(IgnorePath("/healthcheck"))
	})

	mockGetRequest(t, logmiddleware, "https://example.com/healthcheck")
	assert.Empty(t, logbuffer.String())

	// now with a different path
	mockGetRequest(t, logmiddleware, "https://example.com/other")
	assert.NotEmpty(t, logbuffer.String())
}
