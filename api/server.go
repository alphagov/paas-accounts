package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/alphagov/paas-accounts/database"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Config struct {
	DB                *database.DB
	BasicAuthUsername string
	BasicAuthPassword string
	LogWriter         io.Writer
}

type EchoCustomValidator struct {
	validator *validator.Validate
}

func (ecv *EchoCustomValidator) Validate(i interface{}) error {
	return ecv.validator.Struct(i)
}

// New creates a new server. Use ListenAndServe to start accepting connections.
func NewServer(config Config) *echo.Echo {

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(basicAuth(config.BasicAuthUsername, config.BasicAuthPassword))

	if config.LogWriter != nil {
		e.Logger.SetOutput(config.LogWriter)
	}

	e.Validator = &EchoCustomValidator{validator: validator.New()}

	e.GET("/", status)
	e.POST("/agreements", PostAgreementsHandler(config.DB))
	e.POST("/agreements/", PostAgreementsHandler(config.DB))
	e.PUT("/documents/:name", PutDocumentHandler(config.DB))
	e.GET("/documents/:name", GetDocumentHandler(config.DB))
	e.GET("/users/:uuid", GetUserHandler(config.DB))
	e.GET("/users", GetUsersHandler(config.DB))
	e.GET("/users/", GetUsersHandler(config.DB))
	e.POST("/users", PostUserHandler(config.DB))
	e.POST("/users/", PostUserHandler(config.DB))
	e.PATCH("/users/:uuid", PatchUserHandler(config.DB))
	e.GET("/users/:uuid/documents", GetUserDocumentsHandler(config.DB))

	e.HTTPErrorHandler = ErrorHandler

	return e
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
