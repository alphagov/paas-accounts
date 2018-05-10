package api

import (
	"net/http"
	"time"

	"github.com/alphagov/paas-accounts/database"
	"github.com/labstack/echo"
)

func PutDocumentHandler(db *database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var document database.Document
		err := c.Bind(&document)
		if err != nil {
			return err
		}

		document.Name = c.Param("name")
		document.ValidFrom = time.Now()
		err = db.PutDocument(document)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}
