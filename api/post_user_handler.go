package api

import (
	"fmt"
	"github.com/alphagov/paas-accounts/database"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"net/http"
)

func PostUserHandler(db *database.DB) echo.HandlerFunc {
		return func(c echo.Context) error {
			var user database.User
			err := c.Bind(&user)
			if err != nil {
				return err
			}

			err = c.Validate(user)
			if err != nil {
				valerr := err.(validator.ValidationErrors)
				s := fmt.Sprint(valerr)
				return c.JSON(http.StatusBadRequest, s)
			}

			// No two users can have the same email
			_, err = db.GetUserByEmail(*user.Email)
			if err == nil {
				return c.NoContent(http.StatusBadRequest)
			}

			_, err = db.GetUser(user.UUID)
			if err == nil {
				// Return a 201 so we're not leaking what does and doesn't exist
				return c.NoContent(http.StatusCreated)
			}

			err = db.PostUser(user)
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusCreated)
	}
}
