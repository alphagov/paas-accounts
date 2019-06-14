package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

type PatchRequest struct {
	Email    *string `json:"user_email" validate:"omitempty,email"`
	Username string  `json:"username" validate:"required"`
}

func PatchUserHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload PatchRequest
		err := c.Bind(&payload)
		if err != nil {
			return err
		}

		err = c.Validate(payload)
		if err != nil {
			valerr := err.(validator.ValidationErrors)
			s := fmt.Sprint(valerr)
			return c.JSON(http.StatusBadRequest, s)
		}

		user, err := db.GetUser(c.Param("uuid"))
		if err != nil {
			if err == database.ErrUserNotFound {
				return c.JSON(http.StatusNotFound, err)
			}

			return err
		}

		if user.Username == nil {
			user.Username = &payload.Username
		}

		user.Email = payload.Email

		err = db.PatchUser(user)
		if err != nil {
			return err
		}

		updateduser, err := db.GetUser(c.Param("uuid"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusAccepted, updateduser)
	}
}
