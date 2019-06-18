package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("GetDocumentHandler", func() {
	var (
		db     *database.DB
		tempDB *database.TempDB
	)

	BeforeEach(func() {
		var err error
		tempDB, err = database.NewTempDB()
		Expect(err).ToNot(HaveOccurred())

		db, err = database.NewDB(tempDB.TempConnectionString)
		Expect(err).ToNot(HaveOccurred())

		Expect(db.Init()).To(Succeed())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should get a document", func() {
		input := database.Document{
			Name:      "one",
			Content:   "content one",
			ValidFrom: time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC),
		}
		Expect(db.PutDocument(input)).To(Succeed())

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/documents/:name")
		ctx.SetParamNames("name")
		ctx.SetParamValues(input.Name)

		handler := GetDocumentHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`{
			"name": "one",
			"content": "content one",
			"valid_from": "2001-01-01T01:01:01Z"
		}`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})

	It("should return a 404 for document that doesn't exist", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/documents/:name")
		ctx.SetParamNames("name")
		ctx.SetParamValues("one")

		handler := GetDocumentHandler(db)
		Expect(handler(ctx)).To(BeAssignableToTypeOf(NotFoundError{}))
	})
})
