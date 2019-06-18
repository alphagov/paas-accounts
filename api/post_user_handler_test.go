package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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


		err = db.PostUser(database.User{
			UUID:     "11111111-1111-1111-1111-111111111111",
			Email:    strPoint("jeff@jefferson.com"),
			Username: strPoint("jeff@jefferson.com"),
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should add a new user", func() {
		user := database.User{
			UUID:     "00000000-0000-0000-0000-000000000001",
			Email:    strPoint("example@example.com"),
			Username: strPoint("example@example.com"),
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
		Expect(res.Body.String()).To(MatchJSON(`{
			"user_uuid": "00000000-0000-0000-0000-000000000001",
			"user_email": "example@example.com",
			"username": "example@example.com"
		}`))
		Expect(res.Code).To(Equal(http.StatusCreated))

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(user.UUID))
		Expect(userData.Email).To(Equal(user.Email))
	})

	It("should validate the input payload", func() {
		payload := `{"user_uuid": "00000000-0000-0000-0000-000000000001", "user_email": "e@ma.il", "wrong_username_field": "email@example.com"}`
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
		Expect(handler(ctx)).To(BeAssignableToTypeOf(ValidationError{}))
	})

	It("should return BadRequest if a user with the same username exists", func(){
		payload := `{
			"user_uuid": "00000000-0000-0000-0000-000000000001", 
			"wrong_email_field": "e@ma.il", 
			"username": "jeff@jefferson.com"
		}`
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
