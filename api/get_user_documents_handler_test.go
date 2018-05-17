package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("GetUserDocumentsHandler", func() {
	var (
		db                       *database.DB
		tempDB                   *database.TempDB
		user                     database.User
		documentOne, documentTwo database.Document
		agreement                database.Agreement
	)

	BeforeEach(func() {
		var err error
		tempDB, err = database.NewTempDB()
		Expect(err).ToNot(HaveOccurred())

		db, err = database.NewDB(tempDB.TempConnectionString)
		Expect(err).ToNot(HaveOccurred())

		Expect(db.Init()).To(Succeed())

		user = database.User{
			UUID: "00000000-0000-0000-0000-000000000001",
		}
		Expect(db.PutUser(user)).To(Succeed())

		documentOne = database.Document{
			Name:      "document-one",
			Content:   "content one",
			ValidFrom: time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC),
		}
		Expect(db.PutDocument(documentOne)).To(Succeed())

		documentTwo = database.Document{
			Name:      "document-two",
			Content:   "content two",
			ValidFrom: time.Date(2002, 2, 2, 2, 2, 2, 0, time.UTC),
		}
		Expect(db.PutDocument(documentTwo)).To(Succeed())

		agreement = database.Agreement{
			UserUUID:     user.UUID,
			DocumentName: documentOne.Name,
			Date:         documentOne.ValidFrom,
		}
		Expect(db.PutAgreement(agreement)).To(Succeed())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should get all documents", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid/documents")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(user.UUID)

		handler := GetUserDocumentsHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`[
			{
				"name": "document-one",
				"content": "content one",
				"valid_from": "` + documentOne.ValidFrom.Format(time.RFC3339) + `",
				"agreement_date": "` + agreement.Date.Format(time.RFC3339) + `"
			},
			{
				"name": "document-two",
				"content": "content two",
				"valid_from": "` + documentTwo.ValidFrom.Format(time.RFC3339) + `",
				"agreement_date": null
			}
		]`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})

	It("should get all unagreed documents", func() {
		q := url.Values{
			"agreed": []string{"false"},
		}
		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid/documents")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(user.UUID)

		handler := GetUserDocumentsHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`[
			{
				"name": "document-two",
				"content": "content two",
				"valid_from": "` + documentTwo.ValidFrom.Format(time.RFC3339) + `",
				"agreement_date": null
			}
		]`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})

	It("should return unagreed documents when user does not exist", func() {
		unknownUserUUID := "00000000-0000-0000-0000-000000000005"

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid/documents")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(unknownUserUUID)

		handler := GetUserDocumentsHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`[
			{
				"name": "document-one",
				"content": "content one",
				"valid_from": "` + documentOne.ValidFrom.Format(time.RFC3339) + `",
				"agreement_date": null
			},
			{
				"name": "document-two",
				"content": "content two",
				"valid_from": "` + documentTwo.ValidFrom.Format(time.RFC3339) + `",
				"agreement_date": null
			}
		]`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))

	})
})
