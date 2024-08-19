package slogecho

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type requestLoggerConfigModifier func(*middleware.RequestLoggerConfig)

type Middleware struct {
	logger                       *slog.Logger
	extraAttrFuncs               [](func(echo.Context) []slog.Attr)
	requestLoggerConfigModifiers []requestLoggerConfigModifier
	filters                      []Filter
}

func (m Middleware) WithExtraAttrFunc(f func(echo.Context) []slog.Attr) Middleware {
	m.extraAttrFuncs = append(m.extraAttrFuncs, f)
	return m
}

func (m Middleware) WithRequestLoggerConfigModifier(modifiers ...requestLoggerConfigModifier) Middleware {
	m.requestLoggerConfigModifiers = append(m.requestLoggerConfigModifiers, modifiers...)
	return m
}

func (m Middleware) WithFilter(filters ...Filter) Middleware {
	m.filters = append(m.filters, filters...)
	return m
}

func (m Middleware) logValuesFunc(c echo.Context, v middleware.RequestLoggerValues) error {
	for _, filter := range m.filters {
		if !filter(c) {
			return nil
		}
	}
	slogLevel := slog.LevelInfo
	requestAttrs := []slog.Attr{
		slog.String("method", v.Method),
		slog.String("uri", v.URI),
	}
	responseAttrs := []slog.Attr{
		slog.Int("status", v.Status),
	}
	if v.Error != nil {
		responseAttrs = append(responseAttrs, slog.Any("error", v.Error))
	}
	if v.Error != nil || v.Status >= 500 {
		slogLevel = slog.LevelError
	}
	extraAttrs := []slog.Attr{}
	for _, f := range m.extraAttrFuncs {
		extraAttrs = append(extraAttrs, f(c)...)
	}
	attrs := append(
		[]slog.Attr{
			{
				Key:   "request",
				Value: slog.GroupValue(requestAttrs...),
			},
			{
				Key:   "response",
				Value: slog.GroupValue(responseAttrs...),
			},
		},
		extraAttrs...,
	)
	m.logger.LogAttrs(c.Request().Context(), slogLevel, "REQUEST", attrs...)
	return nil
}

func (m Middleware) EchoMiddleware() echo.MiddlewareFunc {
	requestLoggerConfig := middleware.RequestLoggerConfig{
		LogMethod:     true,
		LogURI:        true,
		LogStatus:     true,
		LogError:      true,
		HandleError:   true,
		LogValuesFunc: m.logValuesFunc,
	}

	for _, modifier := range m.requestLoggerConfigModifiers {
		modifier(&requestLoggerConfig)
	}

	return middleware.RequestLoggerWithConfig(requestLoggerConfig)
}

func New(logger *slog.Logger) Middleware {
	return Middleware{
		logger: logger,
	}
}
