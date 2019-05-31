package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("GetUsersHandler", func() {
	var (
		db     *database.DB
		tempDB *database.TempDB
		user1  database.User
		user2  database.User
		user3  database.User
	)

	BeforeEach(func() {
		var err error
		tempDB, err = database.NewTempDB()
		Expect(err).ToNot(HaveOccurred())

		db, err = database.NewDB(tempDB.TempConnectionString)
		Expect(err).ToNot(HaveOccurred())

		Expect(db.Init()).To(Succeed())

		user1 = database.User{
			UUID:  "00000000-0000-0000-0000-000000000001",
			Email: "example1@example.com",
		}
		Expect(db.PutUser(user1)).To(Succeed())
		user2 = database.User{
			UUID:  "00000000-0000-0000-0000-000000000002",
			Email: "example2@example.com",
		}
		Expect(db.PutUser(user2)).To(Succeed())
		user3 = database.User{
			UUID:  "00000000-0000-0000-0000-000000000003",
			Email: "example3@example.com",
		}
		Expect(db.PutUser(user3)).To(Succeed())

	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should get users by uuids", func() {
		q := url.Values{
			"guids": []string{"00000000-0000-0000-0000-000000000001,00000000-0000-0000-0000-000000000002,00000000-0000-0000-0000-000000000003"},
		}
		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users")

		handler := GetUsersHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`{
			"users": [{
				"user_uuid": "00000000-0000-0000-0000-000000000001",
				"user_email": "example1@example.com"
			},
			{
				"user_uuid": "00000000-0000-0000-0000-000000000002",
				"user_email": "example2@example.com"
			},
			{
				"user_uuid": "00000000-0000-0000-0000-000000000003",
				"user_email": "example3@example.com"
			}]
		  }`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})

	It("should get a user by email", func() {
		q := url.Values{}
		q.Set("email", "example3@example.com")
		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users")

		handler := GetUsersHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`{
			"users": [{
				"user_uuid": "00000000-0000-0000-0000-000000000003",
				"user_email": "example3@example.com"
			}]
		  }`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})

	It("should return a 400 if the email or guids query param isn't provided", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/users")

		handler := GetUsersHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body).To(MatchJSON(`"Requires either a guids or email query param"`))
		Expect(res.Code).To(Equal(http.StatusBadRequest))
		Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
	})
})
