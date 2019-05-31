package api

import (
	"net/http"
	"strings"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

func GetUsersHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type Users struct {
			Users []database.User `json:"users"`
		}

		params := c.QueryParams()
		users := Users{}

		if len(params["guids"]) > 0 {
			users.Users, _ = db.GetUsersByUUID(strings.Split(params["guids"][0], ","))
			return c.JSON(http.StatusOK, users)
		}

		if params.Get("email") != "" {
			user, err := db.GetUserByEmail(params.Get("email"))
			if err != nil {
				return err
			}
			users.Users = append(users.Users, user)
			return c.JSON(http.StatusOK, users)
		}

		return c.JSON(http.StatusBadRequest, "Requires either a guids or email query param")
	}
}
