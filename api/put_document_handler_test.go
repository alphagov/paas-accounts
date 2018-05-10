package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("PutDocumentHandler", func() {
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

	It("should accept a document", func() {
		inputName := "one"
		input := database.Document{
			Content: "content one",
		}

		buf, err := json.Marshal(input)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/documents/:name")
		ctx.SetParamNames("name")
		ctx.SetParamValues(inputName)

		handler := PutDocumentHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(BeEmpty())
		Expect(res.Code).To(Equal(http.StatusCreated))

		document, err := db.GetDocument(inputName)
		Expect(err).ToNot(HaveOccurred())
		Expect(document.Name).To(Equal(inputName))
		Expect(document.Content).To(Equal(input.Content))
		Expect(document.ValidFrom).To(BeTemporally("~", time.Now(), time.Minute))
	})
})
