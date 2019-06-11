package api_test

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("PostUserHandler", func() {
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

	It("should add a new user", func() {
		user := database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "example@example.com",
		}

		buf, err := json.Marshal(user)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()

		server := NewServer(Config{
			DB:                db,
			BasicAuthUsername: "jeff",
			BasicAuthPassword: "jefferson",
			LogWriter:         GinkgoWriter,
		})

		ctx := server.AcquireContext()
		ctx.Reset(req, res)
		ctx.SetPath("/users/")

		handler := PostUserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(BeEmpty())
		Expect(res.Code).To(Equal(http.StatusCreated))

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(user.UUID))
		Expect(userData.Email).To(Equal(user.Email))
	})

	It("should validate the input payload", func() {
		payload := `{"user_uuid": "00000000-0000-0000-0000-000000000001", "wrong_email_field": "e@ma.il"}`
		buf := []byte(payload)

		req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()

		server := NewServer(Config{
			DB:                db,
			BasicAuthUsername: "jeff",
			BasicAuthPassword: "jefferson",
			LogWriter:         GinkgoWriter,
		})

		ctx := server.AcquireContext()
		ctx.Reset(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues("00000000-0000-0000-0000-000000000001")

		handler := PostUserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Code).To(Equal(http.StatusBadRequest))
	})
})
