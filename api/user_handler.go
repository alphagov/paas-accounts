package api

import (
	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
	"net/http"
)

func UserHandler(db *database.DB) echo.HandlerFunc {
		return func(c echo.Context) error {
			var userUpdate database.User
			err := c.Bind(&userUpdate)
			if err != nil {
				return err
			}

			userUpdate.UUID = c.Param("uuid")

			req := c.Request()

			if req.Method == "PATCH" {
				err = db.PatchUser(userUpdate)
				if err != nil {
					return err
				}

			}

			if req.Method == "PUT" {
				err = db.PutUser(userUpdate)
				if err != nil {
					return err
				}
			}

			return c.JSON(http.StatusAccepted, userUpdate)
	}
}
