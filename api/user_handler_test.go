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

var _ = Describe("PatchUserHandler", func() {
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

		user := database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "example@example.com",
		}
		Expect(db.PutUser(user)).To(Succeed())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should update an existing user's email", func() {
		userUUID := "00000000-0000-0000-0000-000000000001"
		user := database.User{
			Email: "newexample@example.com",
		}

		buf, err := json.Marshal(user)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PATCH, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(userUUID)

		handler := UserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(Equal(`{"user_uuid":"00000000-0000-0000-0000-000000000001","user_email":"newexample@example.com"}`))
		Expect(res.Code).To(Equal(http.StatusAccepted))

		user.UUID = userUUID

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(userUUID))
		Expect(userData.Email).To(Equal(user.Email))
	})
})

var _ = Describe("PutUserHandler", func() {
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

	It("should accept an update to a user", func() {
		user := database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "example@example.com",
		}
		Expect(db.PutUser(user)).To(Succeed())

		userUpdate := database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "newexample@example.com",
		}

		buf, err := json.Marshal(userUpdate)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(user.UUID)

		handler := UserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(Equal(`{"user_uuid":"00000000-0000-0000-0000-000000000001","user_email":"newexample@example.com"}`))
		Expect(res.Code).To(Equal(http.StatusAccepted))

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(userUpdate.UUID))
		Expect(userData.Email).To(Equal(userUpdate.Email))
	})

	It("should add a new user", func() {
		user := database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "example@example.com",
		}

		buf, err := json.Marshal(user)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users/:uuid")
		ctx.SetParamNames("uuid")
		ctx.SetParamValues(user.UUID)

		handler := UserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(Equal(`{"user_uuid":"00000000-0000-0000-0000-000000000001","user_email":"example@example.com"}`))
		Expect(res.Code).To(Equal(http.StatusAccepted))

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(user.UUID))
		Expect(userData.Email).To(Equal(user.Email))
	})
})
