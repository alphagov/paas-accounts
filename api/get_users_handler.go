package api

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

func GetUsersHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type Users struct {
			Users []*database.User `json:"users"`
		}

		params := c.QueryParams()
		users := Users{}

		if len(params["uuids"]) > 0 && params["uuids"][0] != "" {

			uuids := strings.Split(params["uuids"][0], ",")
			for _, value := range uuids {
				_, err := uuid.FromString(value)

				if err != nil {
					return c.JSON(http.StatusBadRequest, fmt.Sprintf("bad uuid: %s", value))
				}
			}

			results, err := db.GetUsersByUUID(strings.Split(params["uuids"][0], ","))

			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			users.Users = results
			return c.JSON(http.StatusOK, users)
		}

		email := params.Get("email")
		if email != "" {
			dbUsers, err := db.GetUserByEmail(email)
			if err != nil {

				if err == database.ErrUserNotFound {
					return c.NoContent(http.StatusNotFound)
				}

				return err
			}

			if len(dbUsers) > 0 {
				users.Users = dbUsers
			} else {
				users.Users = []*database.User{}
			}
			return c.JSON(http.StatusOK, users)
		}

		return c.JSON(http.StatusBadRequest, "Requires either a uuids or email query param")
	}
}
