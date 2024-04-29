package services

import (
	"context"
	"fmt"
	"github.com/djedjethai/go-oauth2-openid/server"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/repository"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
	"net/http"
)

// TODO redis must be set as LRU,
// imagine saving datas to temporaryStore fail, then data in redis would never be deleted!!!!

// List of the allowed services
var apiWhiteList = map[string]string{
	// "888888":     "88888888",
	"order":       "order",
	"brokerSvc":   "brokerSvc",
	"registrySvc": "registrySvc",
}

type IAuthenticationService interface {
	AuthorizeService(w http.ResponseWriter, r *http.Request) error
	SignoutService(r *http.Request, cookie *http.Cookie) e.IError
	SignupService(w http.ResponseWriter, r *http.Request) e.IError
	SigninService(w http.ResponseWriter, r *http.Request) e.IError
	ApiAuthService(w http.ResponseWriter, r *http.Request) e.IError
}

type AuthenticationService struct {
	// repository
	srv   *server.Server
	repos *repository.Repository
}

func NewAuthenticationService(srv *server.Server, r *repository.Repository) IAuthenticationService {
	return &AuthenticationService{srv, r}
}

func (a *AuthenticationService) AuthorizeService(w http.ResponseWriter, r *http.Request) error {

	return a.srv.HandleAuthorizeRequest(w, r)

}

// ApiAuthService is the authentication endpoint for the API's servers
func (a *AuthenticationService) ApiAuthService(w http.ResponseWriter, r *http.Request) e.IError {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("ApiAuthService - hit handler")

	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg("ApiAuthService - parse form failed")
			return e.NewCustomHTTPStatus(e.StatusInternalServerError, "auth/v1/apiauth", err.Error())
		}
	}

	clientID := r.Form.Get("client_id")
	if clientID == "" {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("ApiAuthService - client_id is empty")
		return e.NewCustomHTTPStatus(e.StatusBadRequest)
	}

	// make sure the client's api is allow
	clid, ok := apiWhiteList[clientID]
	if ok {
		// save in db
		svcData := models.APIserverRedisDatas{
			// ServiceID: clid, // will be add with the deserializeAPIServer func
			Path: "apiauth", // usefull ??
		}

		// save all data present in the form, like serviceID(client_id)
		err := a.repos.RedisAPIserverSet(clid, svcData)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("ApiAuthService - saving %s to redis failed", clid))
		}

		// save APIserver in a temporary store for the user to be reconized later on
		// TODO save in APIserver db(if entry already there skip) the server
		key := fmt.Sprintf("LoggedInUserID-%v", clientID)
		err = a.repos.TemporarySet(key, clientID)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("ApiAuthService - saving %s to temporaryStore failed", clientID))
			return e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}

		obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
			Msg("ApiAuthService - exit successfully")

		// the err response is handled within AuthorizeService
		_ = a.AuthorizeService(w, r)

		return nil
	} else {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("ApiAuthService - invalid client_id")
		return e.NewCustomHTTPStatus(e.StatusForbidden)
	}
}

// SignoutService is the signout endpoint for the users
func (a *AuthenticationService) SignoutService(r *http.Request, cookie *http.Cookie) e.IError {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("SignoutService - hit handler")

	jwt := cookie.Value

	// validate the token
	keyID := "theKeyID"
	secretKey := "mySecretKey"
	encoding := "HS256"

	// make sure the user is authenticated
	usrData, err := a.srv.HandleJWTokenGetdata(context.TODO(), r, jwt, keyID, secretKey, encoding)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("SignoutService - user not authenticated")
		return e.NewCustomHTTPStatus(e.StatusUnauthorized)
	}

	userEmail, ok := usrData["sub"]
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("SignoutService - sub unfound")
		return e.NewCustomHTTPStatus(e.StatusForbidden)
	}

	// get user refreshToken from user DB
	userDatas, err := a.repos.UserGetByEmail(userEmail.(string))
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("SignoutService - get %s from db failed", userEmail.(string)))
		return e.NewCustomHTTPStatus(e.StatusForbidden)
	}

	// Delete all tokens using the refreshToken(I could use the access token as well)
	err = a.srv.Manager.RemoveAllTokensByRefreshToken(context.Background(), userDatas.RefreshTK)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg(fmt.Sprintf("SignoutService - remove %s from db failed", userDatas.RefreshTK))
		return e.NewCustomHTTPStatus(e.StatusInternalServerError)
	}

	// TODO do I delete something(the refreshToken and jwtRefresh) in the userAccount ?
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("SignoutService - exit successfully")

	return nil
}

// SignupService is the signup endpoint for the users
func (a *AuthenticationService) SignupService(w http.ResponseWriter, r *http.Request) e.IError {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("SignupService - hit handler")

	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg("SignupService - parse form failed")
			return e.NewCustomHTTPStatus(e.StatusInternalServerError, "auth/v1/signup", err.Error())
		}
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// some logic
	if len(email) < 1 || len(password) < 1 {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("SignupService - email or password missing")
		return e.NewCustomHTTPStatus(e.StatusForbidden)
	} else {

		// make sure the email does not already exist in db
		err := a.repos.UserIsEmailExist(email)
		if err != nil {
			return err.(e.IError)
		}

		user := models.UserRedisDatas{
			Password: password,
			Path:     "signup",
		}

		// save to redis/cache, as we are not sure the jwt will be deliver
		// save all data present in the form, like password(hash), email
		err = a.repos.RedisUserSet(email, user)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("SignupService - set %v to redis failed", email))
			return e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}

		// save user in a temporary store for the user to be reconized later on
		key := fmt.Sprintf("LoggedInUserID-%v", email)
		err = a.repos.TemporarySet(key, email)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("SignupService - set %v to temporaryStore failed", email))
			return e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}

		obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
			Msg("SignupService - exit successfully")

		_ = a.AuthorizeService(w, r)
		return nil
	}
}

// SigninService is the signin endpoint for the users
func (a *AuthenticationService) SigninService(w http.ResponseWriter, r *http.Request) e.IError {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("SigninService - hit handler")

	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg("SigninService - parse form failed")

			return e.NewCustomHTTPStatus(e.StatusInternalServerError, "auth/v1/signin", err.Error())
		}
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if len(email) < 1 || len(password) < 1 {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Msg("SigninService - email or password are missing")

		return e.NewCustomHTTPStatus(e.StatusForbidden)
	} else {

		user := models.UserRedisDatas{
			Password: password,
			Path:     "signin",
		}

		// save to redis/cache, as we are not sure the jwt will be deliver
		// save all data present in the form, like password(hash), email
		err := a.repos.RedisUserSet(email, user)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("SigninService - set %v to redis failed", email))

		}

		// save user in a temporary store for the user to be reconized later on
		key := fmt.Sprintf("LoggedInUserID-%v", email)
		err = a.repos.TemporarySet(key, email)
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("SigninService - set %v to temporaryStore failed", email))

			return e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("SigninService - exit successfully")

	_ = a.AuthorizeService(w, r)
	return nil
}
