package api

import (
	"net/http"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

func GetUserDocumentsHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		allDocuments, err := db.GetDocumentsForUserUUID(c.Param("uuid"))
		if err != nil {
			return InternalServerError{err}
		}

		onlyUnagreed := c.QueryParam("agreed") == "false"
		userDocuments := []database.UserDocument{}
		for _, doc := range allDocuments {
			if onlyUnagreed && doc.AgreementDate != nil {
				continue
			}
			userDocuments = append(userDocuments, doc)
		}

		return c.JSON(http.StatusOK, userDocuments)
	}
}
