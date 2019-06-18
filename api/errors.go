package api

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"net/http"
)

type NotFoundError struct {
	Message string
}

func (err NotFoundError) Error() string {
	return err.Message
}


type InternalServerError struct {
	InternalError error
}

func (err InternalServerError) Error() string {
	return "Something went wrong internally."
}

type ValidationError struct {
	ValidationErrors validator.ValidationErrors
}

func (err ValidationError) Error() string {
	return "A validation error occurred"
}

type messageErrorBody struct {
	Message string `json:"message"`
}

type validationErrorsBody struct {
	ValidationErrors []fieldValidationError `json:"validation-errors"`
}

type fieldValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func ErrorHandler(err error, ctx echo.Context) {
	switch err.(type) {
	case *echo.HTTPError:
		handleEchoHTTPError(err.(*echo.HTTPError), ctx)

	case NotFoundError:
		handleNotFound(err.(NotFoundError), ctx)

	case InternalServerError:
		handleInternalServerError(err.(InternalServerError), ctx)

	case ValidationError:
		handleValidationError(err.(ValidationError), ctx)

	default:
		handleGenericError(err, ctx)
	}
}



func handleEchoHTTPError(err *echo.HTTPError, ctx echo.Context) {
	code := err.Code
	body := messageErrorBody{ Message: fmt.Sprintf("%v", err.Message) }

	ctx.Logger().Error(err)
	ctx.JSON(code, body)
}

func handleNotFound(err NotFoundError, ctx echo.Context) {
	ctx.Logger().Error(err)
	ctx.JSON(http.StatusNotFound, messageErrorBody{ Message: err.Error() })
}

func handleInternalServerError(err InternalServerError, ctx echo.Context) {
	ctx.Logger().Error(err.InternalError)
	ctx.NoContent(http.StatusInternalServerError)
}

func handleValidationError(err ValidationError, ctx echo.Context) {
	body := validationErrorsBody{ ValidationErrors: []fieldValidationError{} }

	for _, field := range err.ValidationErrors {
		fieldErr := fieldValidationError{
			Field: field.Field(),
			Error: field.Tag(),
		}

		body.ValidationErrors = append(body.ValidationErrors, fieldErr)
	}

	ctx.Logger().Error(err)
	ctx.JSON(http.StatusBadRequest, body)
}

func handleGenericError(err error, ctx echo.Context) {
	ctx.Logger().Error(err)
	ctx.JSON(http.StatusInternalServerError, messageErrorBody{ Message: err.Error() })
}
