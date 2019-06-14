package api

import (
	"net/http"

	"github.com/alphagov/paas-accounts/database"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
)

type PatchRequest struct {
	Email    *string `json:"user_email" validate:"omitempty,email"`
	Username string  `json:"username" validate:"required"`
}

var userNotFoundError = NotFoundError{"user not found"}

func PatchUserHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload PatchRequest
		err := c.Bind(&payload)
		if err != nil {
			return InternalServerError{err}
		}

		err = c.Validate(payload)
		if err != nil {
			valerr := err.(validator.ValidationErrors)
			return ValidationError{valerr}
		}

		user, err := db.GetUser(c.Param("uuid"))
		if err != nil {
			if err == database.ErrUserNotFound {
				return userNotFoundError
			}

			return InternalServerError{err}
		}

		if user.Username == nil {
			user.Username = &payload.Username
		}

		user.Email = payload.Email

		err = db.PatchUser(user)
		if err != nil {
			return InternalServerError{err}
		}

		updateduser, err := db.GetUser(c.Param("uuid"))
		if err != nil {
			if err == database.ErrUserNotFound {
				return userNotFoundError
			}

			return InternalServerError{err}
		}

		return c.JSON(http.StatusAccepted, updateduser)
	}
}
