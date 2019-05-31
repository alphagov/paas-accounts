package api_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("GetUserHandler", func() {
	var (
		db     *database.DB
		tempDB *database.TempDB
		user   database.User
	)

	BeforeEach(func() {
		var err error
		tempDB, err = database.NewTempDB()
		Expect(err).ToNot(HaveOccurred())

		db, err = database.NewDB(tempDB.TempConnectionString)
		Expect(err).ToNot(HaveOccurred())

		Expect(db.Init()).To(Succeed())

		user = database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "example@example.com",
		}
		Expect(db.PutUser(user)).To(Succeed())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should get a user", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(user.UUID)

		handler := GetUserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`{
				"user_uuid": "00000000-0000-0000-0000-000000000001",
				"user_email": "example@example.com"
			}`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})

	It("should return an error if the uuid doesn't exist", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues("00000000-0000-0000-0000-000000000002")

		handler := GetUserHandler(db)
		Expect(handler(ctx)).ToNot(Succeed())
		Expect(res.Code).To(Equal(http.StatusOK))
	})

	It("should return an error if the uuid is incorrect", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues("00000000-0000-0000-0000")

		handler := GetUserHandler(db)
		Expect(handler(ctx)).ToNot(Succeed())
		Expect(res.Code).To(Equal(http.StatusOK))
	})
})
