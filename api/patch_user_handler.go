package api

import (
	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
	"net/http"
)

func PatchUserHandler(db *database.DB) echo.HandlerFunc {
		return func(c echo.Context) error {
			var user database.User
			err := c.Bind(&user)
			if err != nil {
				return err
			}

			user.UUID = c.Param("uuid")

			_, err = db.GetUser(c.Param("uuid"))
			if err != nil {
				return c.JSON(http.StatusNotFound, err)
			}

			err = db.PatchUser(user)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusAccepted, user)
	}
}
