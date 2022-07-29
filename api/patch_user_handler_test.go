package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
			UUID:     "00000000-0000-0000-0000-000000000001",
			Email:    strPoint("example@example.com"),
			Username: strPoint("example@example.com"),
		}
		Expect(db.PostUser(user)).To(Succeed())

		user2 := database.User{
			UUID:     "00000000-0000-0000-0000-000000000002",
			Email:    strPoint("example2@example.com"),
			Username: nil,
		}
		Expect(db.PostUser(user2)).To(Succeed())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should update an existing user's email", func() {
		userUUID := "00000000-0000-0000-0000-000000000001"
		user := database.User{
			Email:    strPoint("newexample@example.com"),
			Username: strPoint("newexample@example.com"),
		}

		buf, err := json.Marshal(user)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PATCH, "/", bytes.NewReader(buf))
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
		ctx.SetParamValues(userUUID)

		handler := PatchUserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(Equal(`{"user_uuid":"00000000-0000-0000-0000-000000000001","user_email":"newexample@example.com","username":"example@example.com"}`))
		Expect(res.Code).To(Equal(http.StatusAccepted))

		user.UUID = userUUID

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(userUUID))
		Expect(userData.Email).To(Equal(user.Email))
	})

	It("should update an existing user's username when the username is null", func() {
		userUUID := "00000000-0000-0000-0000-000000000002"
		user := database.User{
			Email:    strPoint("example2@example.com"),
			Username: strPoint("example2@example.com"),
		}

		buf, err := json.Marshal(user)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PATCH, "/", bytes.NewReader(buf))
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
		ctx.SetParamValues(userUUID)

		handler := PatchUserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(Equal(`{"user_uuid":"00000000-0000-0000-0000-000000000002","user_email":"example2@example.com","username":"example2@example.com"}`))
		Expect(res.Code).To(Equal(http.StatusAccepted))

		user.UUID = userUUID

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(userUUID))
		Expect(userData.Email).To(Equal(user.Email))
		Expect(userData.Username).To(Equal(user.Username))
	})

	It("should not update an existing user's username when the username is not null", func() {
		userUUID := "00000000-0000-0000-0000-000000000001"
		user := database.User{
			Email:    strPoint("example@example.com"),
			Username: strPoint("donotupdate"),
		}

		buf, err := json.Marshal(user)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PATCH, "/", bytes.NewReader(buf))
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
		ctx.SetParamValues(userUUID)

		handler := PatchUserHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(Equal(`{"user_uuid":"00000000-0000-0000-0000-000000000001","user_email":"example@example.com","username":"example@example.com"}`))
		Expect(res.Code).To(Equal(http.StatusAccepted))

		user.UUID = userUUID

		userData, err := db.GetUser(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(userData.UUID).To(Equal(userUUID))
		Expect(userData.Email).To(Equal(user.Email))
		Expect(userData.Username).ToNot(Equal(user.Username))
	})
})
