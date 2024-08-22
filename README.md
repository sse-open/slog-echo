# slog-echo

Simple [echo](https://github.com/labstack/echo) middleware to log http requests using [slog](https://pkg.go.dev/log/slog).


## Usage

```go
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	slogecho "github.com/sse-open/slog-echo"
)

func main() {
	e := echo.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	e.Use(slogecho.New(logger).
		WithFilter(slogecho.IgnorePath("/healthcheck")).
		EchoMiddleware())

	e.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "I'm aliver!")
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.Logger.Fatal(e.Start(":5000"))
}

```


## License notices
MIT License. Copyright (c) 2024 Star Stable Entertainment AB

Parts of this code was forked and/or inspired by https://github.com/samber/slog-echo ( MIT License. Copyright (c) 2023 Samuel Berthe )
