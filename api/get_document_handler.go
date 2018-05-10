package api

import (
	"net/http"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

var ErrDocumentNotFound = echo.NewHTTPError(http.StatusNotFound, "document not found")

func GetDocumentHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		document, err := db.GetDocument(c.Param("name"))
		if err == database.ErrDocumentNotFound {
			return ErrDocumentNotFound
		} else if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, document)
	}
}
