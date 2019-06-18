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
			return InternalServerError{err}
		}

		_, err = db.GetUser(agreement.UserUUID)
		if err != nil {
			err = db.PostUser(database.User{
				UUID: agreement.UserUUID,
			})

			if err != nil {
				return InternalServerError{err}
			}
		}

		agreement.Date = time.Now()
		err = db.PutAgreement(agreement)
		if err != nil {
			return InternalServerError{err}
		}

		return c.NoContent(http.StatusCreated)
	}
}
