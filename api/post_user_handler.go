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

			// No two users can have the same username
			_, err = db.GetUserByUsername(*user.Username)
			if err == nil {
				return c.NoContent(http.StatusBadRequest)
			}

			_, err = db.GetUser(user.UUID)
			if err == nil {
				return c.NoContent(http.StatusConflict)
			}

			err = db.PostUser(user)
			if err != nil {
				return err
			}

			createdUser, err := db.GetUser(user.UUID)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusCreated, createdUser)
	}
}
