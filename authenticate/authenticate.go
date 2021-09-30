package authenticate

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/blaise-cawi-portal/busapi"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

const (
	JWT_TOKEN_KEY           = "jwt_token"
	NO_ACCESS_CODE_ERR      = "Enter an access code"
	INVALID_LENGTH_ERR      = "Enter a 12-character access code"
	NOT_RECOGNISED_ERR      = "Access code not recognised. Enter the code again"
	INTERNAL_SERVER_ERR     = "We were unable to process your request, please try again"
	ISSUER                  = "social-surveys-web-portal"
)

var expirationTime = "2h"

// var expirationTime = "30s"

//Generate mocks by running "go generate ./..."
//go:generate mockery --name AuthInterface
type AuthInterface interface {
	AuthenticatedWithUac(*gin.Context)
	Login(*gin.Context, sessions.Session)
	Logout(*gin.Context, sessions.Session)
	HasSession(*gin.Context) (bool, *UACClaims)
	NotAuthWithError(*gin.Context, string)
}

type Auth struct {
	BusApi        busapi.BusApiInterface
	JWTCrypto     JWTCryptoInterface
	CSRFSecret    string
}

func (auth *Auth) AuthenticatedWithUac(context *gin.Context) {
	session := sessions.Default(context)
	jwtToken := session.Get(JWT_TOKEN_KEY)

	if jwtToken == nil {
		auth.notAuth(context)
		return
	}

	_, err := auth.JWTCrypto.DecryptJWT(jwtToken)
	if err != nil {
		log.Println(err)
		auth.notAuth(context)
		return
	}
	context.Next()
}

func (auth *Auth) HasSession(context *gin.Context) (bool, *UACClaims) {
	session := sessions.Default(context)
	jwtToken := session.Get(JWT_TOKEN_KEY)

	if jwtToken == nil {
		return false, nil
	}

	claim, err := auth.JWTCrypto.DecryptJWT(jwtToken)
	if err != nil || claim == nil {
		return false, nil
	}
	return true, claim
}

func (auth *Auth) Login(context *gin.Context, session sessions.Session) {
	uac := context.PostForm("uac")
	uac = strings.ReplaceAll(uac, " ", "")

	if uac == "" {
		auth.NotAuthWithError(context, NO_ACCESS_CODE_ERR)
		return
	}
	if len(uac) <= 11 || len(uac) >= 13 {
		auth.NotAuthWithError(context, INVALID_LENGTH_ERR)
		return
	}

	uacInfo, err := auth.BusApi.GetUacInfo(uac)
	if err != nil || uacInfo.InstrumentName == "" || uacInfo.CaseID == "" {
		log.Println(err)
		log.Printf("Instrument: %s\n", uacInfo.InstrumentName)
		log.Printf("Case: %s\n", uacInfo.CaseID)
		auth.NotAuthWithError(context, NOT_RECOGNISED_ERR)
		return
	}

	signedToken, err := auth.JWTCrypto.EncryptJWT(uac, &uacInfo)
	if err != nil {
		log.Println(err)
		auth.NotAuthWithError(context, INTERNAL_SERVER_ERR)
		return
	}

	session.Set(JWT_TOKEN_KEY, signedToken)
	if err := session.Save(); err != nil {
		log.Println(err)
		auth.NotAuthWithError(context, INTERNAL_SERVER_ERR)
		return
	}

	context.Redirect(http.StatusFound, fmt.Sprintf("/%s/", uacInfo.InstrumentName))
	context.Abort()
}

func (auth *Auth) Logout(context *gin.Context, session sessions.Session) {
	session.Clear()
	err := session.Save()
	if err != nil {
		auth.notAuth(context)
		return
	}
	context.HTML(http.StatusOK, "logout.tmpl", gin.H{})
	context.Abort()
}

func (auth *Auth) notAuth(context *gin.Context) {
	context.Set("csrfSecret", auth.CSRFSecret)
	context.HTML(http.StatusUnauthorized, "login.tmpl", gin.H{"csrf_token": csrf.GetToken(context)})
	context.Abort()
}

func (auth *Auth) NotAuthWithError(context *gin.Context, errorMessage string) {
	context.Set("csrfSecret", auth.CSRFSecret)
	context.HTML(http.StatusUnauthorized, "login.tmpl", gin.H{"error": errorMessage, "csrf_token": csrf.GetToken(context)})
	context.Abort()
}

func Forbidden(context *gin.Context) {
	context.HTML(http.StatusForbidden, "access_denied.tmpl", gin.H{})
	context.Abort()
}

func expirationSeconds() int64 {
	duration, _ := time.ParseDuration(expirationTime)
	return int64(duration.Seconds())
}
