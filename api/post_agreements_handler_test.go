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

var _ = Describe("PostAgreementsHandler", func() {
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

	It("should accept an agreement", func() {
		document := database.Document{
			Name:      "document-one",
			Content:   "content one",
			ValidFrom: time.Now(),
		}
		Expect(db.PutDocument(document)).To(Succeed())

		user := database.User{
			UUID:     "00000000-0000-0000-0000-000000000001",
			Username: strPoint("example@example.com"),
			Email:    strPoint("example@example.com"),
		}
		Expect(db.PostUser(user)).To(Succeed())

		input := database.Agreement{
			UserUUID:     user.UUID,
			DocumentName: document.Name,
		}

		buf, err := json.Marshal(input)
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewReader(buf))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, res)
		ctx.SetPath("/agreements")

		handler := PostAgreementsHandler(db)
		Expect(handler(ctx)).To(Succeed())
		Expect(res.Body.String()).To(BeEmpty())
		Expect(res.Code).To(Equal(http.StatusCreated))

		agreements, err := db.GetAgreementsForUserUUID(user.UUID)
		Expect(err).ToNot(HaveOccurred())
		Expect(agreements).To(HaveLen(1))
		Expect(agreements[0].UserUUID).To(Equal(input.UserUUID))
		Expect(agreements[0].DocumentName).To(Equal(input.DocumentName))
		Expect(agreements[0].Date).To(BeTemporally("~", time.Now(), time.Minute))
	})
})
