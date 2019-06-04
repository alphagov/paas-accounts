package api

import (
	"github.com/alphagov/paas-accounts/database"
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

			user.UUID = c.Param("uuid")

			_, err = db.GetUser(c.Param("uuid"))
			if err == nil {
				return c.JSON(http.StatusNotFound, "user already exists")
			}

			err = db.PostUser(user)
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusCreated)
	}
}
