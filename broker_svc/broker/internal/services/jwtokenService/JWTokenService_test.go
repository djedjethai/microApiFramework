package jwtokenService

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	obs "gitlab.com/grpasr/common/observability"
	"gitlab.com/grpasr/common/tests"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	key      = "keyID"
	signKey  = "secretKey"
	subject  = "serviceName"
	audience = "serviceDest"
)

var (
	expireInAnHour = time.Now().Add(time.Hour).Unix()
	expiredTime    = time.Now().Add(-time.Hour).Unix()
	validToken     string
	expiredToken   string
	validConfig    *config.Config // means secretKey match const signKey
	invalidConfig  *config.Config // secretKey does not match const signKey
	ui             map[string]string
)

func TestMain(m *testing.M) {
	// Initialize your test context
	setupTests()

	// Run tests
	exitCode := m.Run()

	// Teardown
	teardown()

	// Exit with the appropriate code
	os.Exit(exitCode)

}

func Test_validate_valid_token(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	jwtSVC := NewJWTokenService(validConfig)

	ctx := context.Background()

	data, ce := jwtSVC.JWTokenIsValidToken(ctx, validToken)

	tests.MaybeFail("validate_valid_token", ce, tests.Expect(len(data), 3))
	tests.MaybeFail("validate_valid_token", tests.Expect(data["role"], ui["role"]))
	tests.MaybeFail("validate_valid_token", tests.Expect(data["scope"], ui["scope"]))
	tests.MaybeFail("validate_valid_token", tests.Expect(data["svc"], subject))
}

func Test_validate_expired_token(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	jwtSVC := NewJWTokenService(validConfig)

	ctx := context.Background()

	data, ce := jwtSVC.JWTokenIsValidToken(ctx, expiredToken)

	tests.MaybeFail("http_status", tests.Expect(ce.GetCode(), http.StatusUnauthorized))
	tests.MaybeFail("http_status", tests.Expect(len(ce.GetPayload()), 0))
	tests.MaybeFail("http_status", tests.Expect(len(data), 0))
}

func Test_validate_invalid_secretKey(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	jwtSVC := NewJWTokenService(invalidConfig)

	ctx := context.Background()

	data, ce := jwtSVC.JWTokenIsValidToken(ctx, validToken)

	tests.MaybeFail("http_status", tests.Expect(ce.GetCode(), http.StatusForbidden))
	tests.MaybeFail("http_status", tests.Expect(len(ce.GetPayload()), 0))
	tests.MaybeFail("http_status", tests.Expect(len(data), 0))

}

func Test_validate_invalid_info_token(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	ag := NewJWTAccessGenerate()

	ii := make(map[string]string)
	ii["nameInv"] = "robert"
	ii["scopeInv"] = "read, openid"
	ii["roleInv"] = "APIserver"

	invalidInfoToken, err := ag.GenerateOpenidJWToken(expireInAnHour, ii)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	jwtSVC := NewJWTokenService(validConfig)

	ctx := context.Background()

	data, ce := jwtSVC.JWTokenIsValidToken(ctx, invalidInfoToken)

	tests.MaybeFail("http_status", tests.Expect(ce.GetCode(), http.StatusForbidden))
	tests.MaybeFail("http_status", tests.Expect(ce.Error(), `403 : Request forbidden, Comment: token contains invalid data`))
	tests.MaybeFail("http_status", tests.Expect(len(ce.GetPayload()), 0))
	tests.MaybeFail("http_status", tests.Expect(len(data), 0))
}

func Test_validate_empty_info_token(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	ag := NewJWTAccessGenerate()

	ii := make(map[string]string)
	ii["name"] = "robert"
	ii["scope"] = ""
	ii["role"] = "APIserver"

	invalidInfoToken, err := ag.GenerateOpenidJWToken(expireInAnHour, ii)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	jwtSVC := NewJWTokenService(validConfig)

	ctx := context.Background()

	data, ce := jwtSVC.JWTokenIsValidToken(ctx, invalidInfoToken)

	tests.MaybeFail("validate_valid_token", ce, tests.Expect(len(data), 3))
	tests.MaybeFail("validate_valid_token", tests.Expect(data["role"], ii["role"]))
	tests.MaybeFail("validate_valid_token", tests.Expect(data["scope"], ii["scope"]))
	tests.MaybeFail("validate_valid_token", tests.Expect(data["svc"], subject))
}

func setupTests() {

	obs.SetObservabilityFacade("broker_svc")

	ag := NewJWTAccessGenerate()

	ui = make(map[string]string)
	ui["name"] = "robert"
	ui["scope"] = "read, openid"
	ui["role"] = "APIserver"

	var err error
	validToken, err = ag.GenerateOpenidJWToken(expireInAnHour, ui)
	if err != nil {
		fmt.Println("error created validToken")
	}

	expiredToken, err = ag.GenerateOpenidJWToken(expiredTime, ui)
	if err != nil {
		fmt.Println("error created expiredToken")
	}

	// set configs
	os.Setenv("GOENV", "localhost")
	os.Setenv("CONFIG_FILE_PATH", "../../../configs/v1/servicesAddr/")
	os.Setenv("PATH_TO_TLS", "../../../configs/v1/certificates/")

	// create a config with the valid signedKey
	os.Setenv("SIGNED_KEY", "secretKey")
	validConfig, err = config.SetConfigs()
	if err != nil {
		fmt.Println("error set config: ", err)
	}

	// create a config with an invalid signedKey
	os.Setenv("SIGNED_KEY", "secretKeyInvalid")
	invalidConfig, err = config.SetConfigs()
	if err != nil {
		fmt.Println("error set config: ", err)
	}
}

func teardown() {}

// JWTAccessGenerate generate the jwt access token
type JWTAccessGenerate struct {
	signedKeyID  string // identifiant refering to the SignedKey
	signedKey    []byte // secret key
	signedMethod jwt.SigningMethod
}

func NewJWTAccessGenerate() *JWTAccessGenerate {
	na := &JWTAccessGenerate{}

	na.signedKeyID = key
	na.signedKey = []byte(signKey)
	na.signedMethod = jwt.SigningMethodHS256
	return na
}

func (a *JWTAccessGenerate) GenerateOpenidJWToken(expiresAt int64, ui map[string]string) (string, error) {

	claims := &JWTAccessClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  audience,
			Subject:   subject,
			ExpiresAt: expiresAt,
		},
		UserInfo: ui,
	}

	token := jwt.NewWithClaims(a.signedMethod, claims)
	if a.signedKeyID != "" {
		token.Header["kid"] = a.signedKeyID
	}
	var key interface{}
	key = a.signedKey

	access, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return access, nil
}

type JWTAccessClaims struct {
	jwt.StandardClaims
	UserInfo map[string]string `json:"openidInfo"`
	// AccessToken  string            `json:"accessToken"`
}

// Valid claims verification
func (a *JWTAccessClaims) Valid() error {
	if time.Unix(a.ExpiresAt, 0).Before(time.Now()) {
		// if a.ExpiresAt < oneMonthAgo {
		return fmt.Errorf("error invalid access token")
	}
	return nil
}
