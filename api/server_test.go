package api_test

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var _ = Describe("Server", func() {

	var (
		shutdownServer        context.CancelFunc
		db                    *database.DB
		tempDB                *database.TempDB
		server                *echo.Echo
		addr                  string
		cleanShutdownComplete chan struct{}
		basicUsername         = "jeff"
		basicPassword         = "jefferson"
	)

	BeforeEach(func() {
		var err error
		tempDB, err = database.NewTempDB()
		Expect(err).ToNot(HaveOccurred())

		db, err = database.NewDB(tempDB.TempConnectionString)
		Expect(err).ToNot(HaveOccurred())

		Expect(db.Init()).To(Succeed())

		var ctx context.Context
		ctx, shutdownServer = context.WithCancel(context.Background())

		server = NewServer(Config{
			DB:                db,
			BasicAuthUsername: basicUsername,
			BasicAuthPassword: basicPassword,
			LogWriter:         GinkgoWriter,
		})
		cleanShutdownComplete = make(chan struct{})
		go func() {
			Expect(ListenAndServe(ctx, server, "0.0.0.0:0")).To(Succeed())
			close(cleanShutdownComplete)
		}()
		Eventually(func() net.Listener {
			return server.Listener
		}).ShouldNot(BeNil())
		addr = server.Listener.Addr().String()
	})

	AfterEach(func() {
		shutdownServer()
		Eventually(cleanShutdownComplete, 10*time.Second).Should(BeClosed())
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	DescribeTable("should expose status route to public without basic auth",
		func(method, path string) {
			var body io.Reader
			if method == http.MethodPost {
				body = strings.NewReader("{}")
			}
			url := "http://" + addr + path
			req, err := http.NewRequest(method, url, body)
			Expect(err).ToNot(HaveOccurred())
			client := &http.Client{}
			res, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(200))
		},
		Entry("GET /", "GET", "/"),
	)

	DescribeTable("should not expose routes to public",
		func(method, path string) {
			var body io.Reader
			if method == http.MethodPost {
				body = strings.NewReader("{}")
			}
			url := "http://" + addr + path
			req, err := http.NewRequest(method, url, body)
			Expect(err).ToNot(HaveOccurred())
			client := &http.Client{}
			res, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(401))
			b, err := ioutil.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(b)).To(MatchJSON(`{
				"message": "Unauthorized"
			}`))
		},
		Entry("POST /agreements", "POST", "/agreements"),
		Entry("PUT /documents/:name", "PUT", "/documents/doc-one"),
		Entry("GET /documents/:name", "GET", "/documents/doc-one"),
		Entry("GET /users/569a91c6-7f5d-4dac-82a2-db85cc595c75/documents", "GET", "/users/"),
		Entry("GET /users?guids=569a91c6-7f5d-4dac-82a2-db85cc595c75", "GET", "/users"),
		Entry("PUT /users/:uuid", "PUT", "/users/569a91c6-7f5d-4dac-82a2-db85cc595c75"),
		Entry("PATCH /users/:uuid", "PATCH", "/users/569a91c6-7f5d-4dac-82a2-db85cc595c75"),
	)

	DescribeTable("should allow access with basic auth credentials",
		func(method, path string, responseCode int) {
			var body io.Reader
			if method != http.MethodGet {
				body = strings.NewReader("{}")
			}
			url := "http://" + addr + path
			req, err := http.NewRequest(method, url, body)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
			req.SetBasicAuth(basicUsername, basicPassword)
			client := &http.Client{}
			res, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(responseCode))
		},
		Entry("POST /agreements", "POST", "/agreements", 500),
		Entry("PUT /documents/:name", "PUT", "/documents/doc-one", 500),
		Entry("GET /documents/:name", "GET", "/documents/doc-one", 404),
		Entry("GET /users/:uuid/documents", "GET", "/users/569a91c6-7f5d-4dac-82a2-db85cc595c75/documents", 200),
		Entry("GET /users", "GET", "/users", 400),
		Entry("PUT /users/:uuid", "PUT", "/users/569a91c6-7f5d-4dac-82a2-db85cc595c75", 202),
		Entry("PATCH /users/:uuid", "PATCH", "/users/569a91c6-7f5d-4dac-82a2-db85cc595c75", 202),
	)

	Describe("ErrorHandler", func() {
		It("should return all errors as json", func() {
			req := httptest.NewRequest(echo.GET, "/", nil)
			res := httptest.NewRecorder()

			e := echo.New()
			e.Logger.SetOutput(GinkgoWriter)

			ctx := e.NewContext(req, res)
			ctx.SetPath("/")

			err := errors.New("BANG")
			ErrorHandler(err, ctx)
			Expect(res.Body).To(MatchJSON(`{
				"message": "` + err.Error() + `"
			}`))
			Expect(res.Code).To(Equal(http.StatusInternalServerError))
			Expect(res.Header().Get("Content-Type")).To(Equal(echo.MIMEApplicationJSONCharsetUTF8))
		})

	})

})
