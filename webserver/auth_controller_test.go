package webserver_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/ONSdigital/blaise-cawi-portal/authenticate"
	"github.com/ONSdigital/blaise-cawi-portal/authenticate/mocks"
	"github.com/ONSdigital/blaise-cawi-portal/busapi"
	"github.com/ONSdigital/blaise-cawi-portal/webserver"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	csrf "github.com/srbry/gin-csrf"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Controller", func() {
	var (
		httpRouter  *gin.Engine
		mockAuth    = &mocks.AuthInterface{}
		csrfManager = &csrf.DefaultCSRFManager{
			Secret:      "fwibble",
			SessionName: "session",
		}
		authController = &webserver.AuthController{
			Auth:        mockAuth,
			CSRFManager: csrfManager,
			UacKind:     "uac",
		}
		instrumentName  = "foobar"
		caseID          = "fizzbuzz"
		observedLogs    *observer.ObservedLogs
		observedZapCore zapcore.Core
		config          = &webserver.Config{UacKind: "uac16"}
	)

	BeforeEach(func() {
		observedZapCore, observedLogs = observer.New(zap.InfoLevel)
		observedLogger := zap.New(observedZapCore)
		observedLogger.Sync()
		csrfManager.ErrorFunc = webserver.CSRFErrorFunc(csrfManager, config, observedLogger)
		httpRouter = gin.Default()
		store := cookie.NewStore([]byte("secret"))
		httpRouter.Use(sessions.SessionsMany([]string{"session", "user_session", "session_validation"}, store))
		httpRouter.LoadHTMLGlob("../templates/*")
		authController.Logger = observedLogger
		authController.AddRoutes(httpRouter)
	})

	AfterEach(func() {
		mockAuth = &mocks.AuthInterface{}
		authController = &webserver.AuthController{Auth: mockAuth, CSRFManager: csrfManager}
	})

	Describe("GET /auth/login", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/auth/login", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("when I access auth/login I am presented with the login template", func() {
			BeforeEach(func() {
				mockAuth.On("HasSession", mock.Anything).Return(false, nil)
			})

			It("returns a 200 response and the login page", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(ContainSubstring(`<span class="btn__inner">Access study`))
			})
		})

		Context("when I access auth/login with an active session", func() {
			JustBeforeEach(func() {
				httpRecorder = httptest.NewRecorder()

				mockAuth.On("HasSession", mock.Anything).Return(true, &authenticate.UACClaims{UacInfo: busapi.UacInfo{
					InstrumentName: instrumentName,
					CaseID:         caseID,
				}}, nil)

				req, _ := http.NewRequest("GET", "/auth/login", nil)
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			It("returns a temporary redirect response", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusTemporaryRedirect))

				header := httpRecorder.Header()["Location"]
				Expect(header).To(Equal([]string{fmt.Sprintf("/%s/", instrumentName)}))
			})
		})
	})

	Describe("POST /auth/login", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			mockAuth.On("Login", mock.Anything, mock.Anything).Return()
		})

		Context("without a CSRF", func() {
			JustBeforeEach(func() {
				httpRecorder = httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/auth/login", nil)
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			It("gives an auth error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusForbidden))
				Expect(httpRecorder.Body.String()).To(ContainSubstring(`Request timed out, please try again`))
			})
		})

		Context("with an invalid CSRF", func() {
			JustBeforeEach(func() {
				httpRecorder = httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/auth/login?_csrf=dalajksdqoosk", nil)
				req.RemoteAddr = "1.1.1.1"
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			It("gives an auth error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusForbidden))
				Expect(httpRecorder.Body.String()).To(ContainSubstring(`Request timed out, please try again`))

				Expect(observedLogs.Len()).To(Equal(1))
				Expect(observedLogs.All()[0].Message).To(Equal("CSRF mismatch"))
				Expect(observedLogs.All()[0].ContextMap()["SourceIP"]).To(Equal("1.1.1.1"))
				Expect(observedLogs.All()[0].Level).To(Equal(zap.InfoLevel))
			})
		})

		Context("with a valid CSRF", func() {
			var csrfToken string

			JustBeforeEach(func() {
				httpRouter.GET("/token", func(context *gin.Context) {
					csrfToken = csrfManager.GetToken(context)
				})

				req1, _ := http.NewRequest("GET", "/token", nil)

				httpRecorder = httptest.NewRecorder()
				httpRouter.ServeHTTP(httpRecorder, req1)

				req2, _ := http.NewRequest("POST", fmt.Sprintf("/auth/login?_csrf=%s", csrfToken), nil)
				req2.Header.Set("Cookie", httpRecorder.Header().Get("Set-Cookie"))
				req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				httpRecorder = httptest.NewRecorder()
				httpRouter.ServeHTTP(httpRecorder, req2)
			})

			It("calls it auth.login", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				mockAuth.AssertNumberOfCalls(GinkgoT(), "Login", 1)
			})
		})

		Context("with an invalid UAC Code", func() {
			var csrfToken string

			JustBeforeEach(func() {
				httpRouter.GET("/token", func(context *gin.Context) {
					csrfToken = csrfManager.GetToken(context)
				})

				req1, _ := http.NewRequest("GET", "/token", nil)

				httpRecorder = httptest.NewRecorder()
				httpRouter.ServeHTTP(httpRecorder, req1)

				httpRecorder = httptest.NewRecorder()
				data := url.Values{
					"uac":   []string{"123"},
					"_csrf": []string{csrfToken},
				}
				req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(data.Encode()))
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			Context("Login with a 12 digit UAC kind", func() {
				BeforeEach(func() {
					authController.UacKind = "uac"
					config.UacKind = "uac"
				})

				It("states a 12-digit access code is required", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusForbidden))
					Expect(httpRecorder.Body.String()).To(ContainSubstring(`Enter your 12-digit access code`))
				})
			})

			Context("Login with a 16 character UAC kind", func() {
				BeforeEach(func() {
					authController.UacKind = "uac16"
					config.UacKind = "uac16"
				})

				It("states a 16-character access code is required", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusForbidden))
					Expect(httpRecorder.Body.String()).To(ContainSubstring(`Enter your 16-character access code`))
				})
			})
		})
	})

	Describe("GET /auth/logout", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			mockAuth.On("Logout", mock.Anything, mock.Anything).Return()
		})

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/auth/logout", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		It("calls it auth.logout", func() {
			mockAuth.AssertNumberOfCalls(GinkgoT(), "Logout", 1)
		})
	})

	Describe("GET /auth/logged-in", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/auth/logged-in", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("when you have an active session", func() {
			BeforeEach(func() {
				mockAuth.On("HasSession", mock.Anything).Return(true, nil)
			})

			It("returns OK", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when you don't have an active session", func() {
			BeforeEach(func() {
				mockAuth.On("HasSession", mock.Anything).Return(false, nil)
			})

			It("returns unauthorised", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("Get /auth/timed-out", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/auth/timed-out", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		It("returns the timed out page", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			body := httpRecorder.Body.String()
			Expect(body).To(ContainSubstring(`Sorry, you need to sign in again`))
			Expect(body).To(ContainSubstring(`This is because you've been inactive for 15 minutes`))
		})
	})
})
