package api

import (
	"net/http"
	"time"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

func PostAgreementsHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var agreement database.Agreement
		err := c.Bind(&agreement)
		if err != nil {
			return err
		}

		_, err = db.GetUser(c.Param("uuid"))
		if err != nil {
			db.PostUser(database.User{
				UUID: agreement.UserUUID,
			})
		}

		agreement.Date = time.Now()
		err = db.PutAgreement(agreement)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}
