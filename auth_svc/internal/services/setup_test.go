package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	oauth2 "github.com/djedjethai/go-oauth2-openid"
	"github.com/djedjethai/go-oauth2-openid/errors"
	"github.com/djedjethai/go-oauth2-openid/manage"
	"github.com/djedjethai/go-oauth2-openid/models"
	"github.com/djedjethai/go-oauth2-openid/server"
	"github.com/djedjethai/go-oauth2-openid/store"
	"github.com/golang-jwt/jwt"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/config"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/repository"
	obs "gitlab.com/grpasr/common/observability"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	conf    *config.Config
	srv     *server.Server
	manager *manage.Manager
	csrv    *httptest.Server
	tsrv    *httptest.Server
	repos   *repository.Repository

	authService   IAuthenticationService
	oauth2Service IOauth2Service
	tokenService  ITokenService

	clientID          = "111111"
	clientSecret      = "11111111"
	codeVerifier      = "s256tests256tests256tests256tests256tests256test"
	s256ChallengeHash string

	// jwt
	keyID          string // should match the on passed in the server
	secretKey      string
	expireInAnHour = time.Now().Add(time.Hour).Unix()
	expiredTime    = time.Now().Add(-time.Hour).Unix()
	ui             map[string]interface{}
)

const (
	// TODO all these credentials should be pass via secrets
	// frontend
	idvar     string = "222222"
	secretvar string = "22222222"
	domainvar string = "http://localhost:80"

	// credential for the preOrder service
	orderID     string = "order"
	orderSecret string = "orderSecret"
	orderDomain string = "http://localhost:50001"

	// broker_svc
	brokerSvcID     string = "brokerSvc"
	brokerSvcSecret string = "brokerSvcSecret"
	brokerSvcDomain string = "http://localhost:8080"

	// registry_svc
	registrySvcID     string = "registrySvc"
	registrySvcSecret string = "registrySvcSecret"
	registrySvcDomain string = "http://localhost:4000"
)

func genCodeChallengeS256(s string) {
	s256 := sha256.Sum256([]byte(s))
	s256ChallengeHash = strings.TrimSpace(base64.URLEncoding.EncodeToString(s256[:]))
}

func TestMain(m *testing.M) {

	// set configs
	conf = config.SetConfig()

	// set observability
	obs.SetObservabilityFacade("auth_svc")
	// obs.Logging.SetLoggingEnvToDevelopment()

	// these must match the passed within the svc
	keyID = "theKeyID"
	secretKey = "mySecretKey"

	genCodeChallengeS256(codeVerifier)

	// set server
	manager = manage.NewDefaultManager()
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	manager.MapClientStorage(registerServices())
	// manager.MapClientStorage(clientStore(csrv.URL, true))

	repos = &repository.Repository{
		ITemporaryStore: repository.NewTmpStore(),
		IUserStore:      repository.NewUserStoreMock(),
		IAPIserverStore: repository.NewAPIserverStoreMock(),
		IRedisStore:     repository.NewRedisMock()}
	srv = server.NewServer(server.NewConfig(), manager)
	srv.SetModeAPI()

	authService = NewAuthenticationService(srv, repos)
	oauth2Service = NewOauth2Service(srv, repos)
	tokenService = NewTokenService(srv, repos)

	// set the handler functions
	srv.SetUserAuthorizationHandler(oauth2Service.UserAuthorizeService)
	srv.SetUserOpenidHandler(oauth2Service.UserOpenidService)
	srv.SetCustomizeTokenPayloadHandler(oauth2Service.UserCustomizeTokenPayloadService)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	// ser QueryParams for all tests
	setQueryParams()

	// exit code
	exitCode := m.Run()

	teardown()

	os.Exit(exitCode)
}

func teardown() {
	return
}

func testServer(t *testing.T, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/authorize":
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			t.Error(err)
		}
	case "/token":
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			t.Error(err)
		}
	}
}

func registerServices() oauth2.ClientStore {
	clientStore := store.NewClientStore()
	// register the frontend
	clientStore.Set(idvar, &models.Client{
		ID:     idvar,
		Secret: secretvar,
		Domain: domainvar,
		UserID: "frontend",
	})

	// TODO seems that if credentials are changed, new credentials are not updated...
	// TODO see what the difference between ID and UserID
	// (are the same with configurations right now)...
	// // register prePost service
	// clientStore.Set(&models.Client{
	// 	ID:     orderID,
	// 	Secret: orderSecret,
	// 	Domain: orderDomain,
	// 	UserID: "order",
	// })

	// register broker_svc
	clientStore.Set(brokerSvcID, &models.Client{
		ID:     brokerSvcID,
		Secret: brokerSvcSecret,
		Domain: brokerSvcDomain,
		UserID: "brokerSvc",
	})

	// // register registry_svc
	// clientStore.Set(&models.Client{
	// 	ID:     registrySvcID,
	// 	Secret: registrySvcSecret,
	// 	Domain: registrySvcDomain,
	// 	UserID: "registrySvc",
	// })

	return clientStore
}

// JWTAccessGenerate generate the jwt access token
type JWTAccessGenerate struct {
	signedKeyID  string // identifiant refering to the SignedKey
	signedKey    []byte // secret key
	signedMethod jwt.SigningMethod
}

func NewJWTAccessGenerate(key, signedKey string) *JWTAccessGenerate {
	na := &JWTAccessGenerate{}

	na.signedKeyID = key
	na.signedKey = []byte(signedKey)
	na.signedMethod = jwt.SigningMethodHS256
	return na
}

// clientID is the serviceID(like 22222), the userID is the user email
func (a *JWTAccessGenerate) GenerateOpenidJWToken(expiresAt int64, ui oauth2.OpenidInfo, clientID, userID string) (string, error) {

	claims := &JWTAccessClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  clientID,
			Subject:   userID,
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

func (a *JWTAccessGenerate) ValidOpenidJWToken(ctx context.Context, tokenString string) error {
	var secretKey = a.signedKey

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		claims := token.Claims.(jwt.MapClaims)
		err := claims.Valid()
		if err != nil {
			return nil, errors.ErrExpiredJWToken
		}

		return secretKey, nil
	})
	if err != nil {
		if err.Error() == "token contains an invalid number of segments" ||
			err.Error() == "signature is invalid" {
			return errors.ErrInvalidJWToken
		}
		return err
	}

	if token.Valid {
		return nil
	} else {
		return errors.ErrInvalidJWToken
	}
}

// GetdataOpenidJWToken return the user's data stored into the JWT
func (a *JWTAccessGenerate) GetdataOpenidJWToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {

	data := make(map[string]interface{})

	var secretKey = a.signedKey

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		claims := token.Claims.(jwt.MapClaims)
		err := claims.Valid()
		if err != nil {
			return nil, errors.ErrExpiredJWToken
		}

		return secretKey, nil
	})
	if err != nil {
		return data, errors.ErrInvalidJWToken
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Access openidInfo claims
		// TODO maybe loop on the claim to get all datas
		data["email"] = claims["sub"]
		data["expiresAt"] = claims["exp"]
		if openidInfo, ok := claims["openidInfo"].(map[string]interface{}); ok {
			for k, v := range openidInfo {
				data[k] = v
			}
		}
	} else {
		return data, errors.ErrInvalidJWToken
	}

	return data, nil

}

type JWTAccessClaims struct {
	jwt.StandardClaims
	UserInfo oauth2.OpenidInfo `json:"openidInfo"`
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
