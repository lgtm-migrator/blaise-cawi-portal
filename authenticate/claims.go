package authenticate

import (
	"strings"

	"github.com/ONSdigital/blaise-cawi-portal/busapi"
	"github.com/golang-jwt/jwt"
)

type UACClaims struct {
	UAC               string `json:"uac"`
	PostcodeValidated bool   `json:"postcode_validated"`
	busapi.UacInfo
	jwt.StandardClaims
}

func (uacClaims *UACClaims) AuthenticatedForInstrument(instrumentName string) bool {
	return strings.ToLower(uacClaims.UacInfo.InstrumentName) == strings.ToLower(instrumentName)
}

func (uacClaims *UACClaims) AuthenticatedForCase(caseID string) bool {
	return strings.ToLower(uacClaims.UacInfo.CaseID) == strings.ToLower(caseID)
}
