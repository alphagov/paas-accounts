package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Config struct {
	DB                *database.DB
	BasicAuthUsername string
	BasicAuthPassword string
	LogWriter         io.Writer
}

// New creates a new server. Use ListenAndServe to start accepting connections.
func NewServer(config Config) *echo.Echo {

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(basicAuth(config.BasicAuthUsername, config.BasicAuthPassword))

	if config.LogWriter != nil {
		e.Logger.SetOutput(config.LogWriter)
	}

	e.GET("/", status)
	e.POST("/agreements", PostAgreementsHandler(config.DB))
	e.PUT("/documents/:name", PutDocumentHandler(config.DB))
	e.GET("/documents/:name", GetDocumentHandler(config.DB))
	e.GET("/users/:uuid/documents", GetUserDocumentsHandler(config.DB))

	e.HTTPErrorHandler = ErrorHandler

	return e
}

func ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := err.Error()
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprintf("%v", he.Message)
	}
	errJSON := struct {
		Message string `json:"message"`
	}{
		Message: msg,
	}
	c.Logger().Error(err)
	c.JSON(code, errJSON)
}

func status(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, map[string]bool{
		"ok": true,
	}, "  ")
}

func ListenAndServe(ctx context.Context, e *echo.Echo, addr string) error {
	ctx, shutdown := context.WithCancel(ctx)

	go func() {
		defer shutdown()
		if err := e.Start(addr); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				e.Logger.Error("listen-and-serve-error", err)
			}
		}
	}()

	// Wait for parent context to get cancelled then drain with a 10s timeout
	<-ctx.Done()
	drainCtx, cancelDrain := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDrain()
	return e.Shutdown(drainCtx)
}

func basicAuth(masterUsername string, masterPassword string) echo.MiddlewareFunc {
	if masterUsername == "" {
		panic("a basic auth username is required")
	}
	if masterPassword == "" {
		panic("a basic auth password is required")
	}
	return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: func(c echo.Context) bool {
			if c.Path() == "/" {
				return true
			}
			return false
		},
		Validator: func(requestUsername, requestPassword string, c echo.Context) (bool, error) {
			if requestUsername == masterUsername && requestPassword == masterPassword {
				return true, nil
			}
			return false, nil
		},
	})
}
