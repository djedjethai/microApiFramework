package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/djedjethai/go-oauth2-openid/server"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/repository"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
	"log"
	"net/http"
	"strings"
	"time"
)

type ITokenService interface {
	RefreshOpenidService(w http.ResponseWriter, r *http.Request, cookie *http.Cookie)
	TokenService(w http.ResponseWriter, r *http.Request, authHeader string) e.IError
	JwtGetdataService(w http.ResponseWriter, r *http.Request, cookie *http.Cookie) (map[string]interface{}, e.IError)
	JwtValidationService(r *http.Request, cookie *http.Cookie) e.IError
	ValidPermissionService(w http.ResponseWriter, r *http.Request) (map[string]interface{}, e.IError)
}

type TokenService struct {
	srv   *server.Server
	repos *repository.Repository
}

func NewTokenService(srv *server.Server, rp *repository.Repository) ITokenService {
	return &TokenService{
		srv:   srv,
		repos: rp,
	}
}

// RefreshOpenidService refresh of the jwt_access
func (t *TokenService) RefreshOpenidService(w http.ResponseWriter, r *http.Request, cookie *http.Cookie) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("RefreshOpenidService - hit service", r)

	jwt := cookie.Value

	// validate the token
	keyID := "theKeyID"
	secretKey := "mySecretKey"
	encoding := "HS256"

	// use this method which returns the data even the jwt is expired
	data, err := t.srv.HandleJWTokenAdminGetdata(context.TODO(), r, jwt, keyID, secretKey, encoding)

	emailOrAPIsvcID, ok := data["sub"]
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("RefreshOpenidService - emailOrAPIsvcID missing")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden))
		return
	}

	role := data["role"].(string)

	switch role {
	case "user":
		user, err := t.repos.UserGetByEmail(emailOrAPIsvcID.(string))
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("RefreshOpenidService - get %v from db failed", emailOrAPIsvcID))
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden))
			return
		}

		r.Header.Set("refresh_token", user.RefreshTK)
		r.Header.Set("jwt_refresh_token", user.RefreshJWT)
		r.Header.Set("jwt_access_token", jwt)

	case "APIserver":
		// NOTE check the serviceID is whiteList ??? overkilled ?
		apiSvc, err := t.repos.APIserverGetByID(emailOrAPIsvcID.(string))
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg(fmt.Sprintf("RefreshOpenidService - get %v from db failed", emailOrAPIsvcID))
			// do something better....
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden))
			return
		}

		r.Header.Set("refresh_token", apiSvc.RefreshTK)
		r.Header.Set("jwt_refresh_token", apiSvc.RefreshJWT)
		r.Header.Set("jwt_access_token", jwt)

	default:
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden))
		return
	}

	err = r.ParseForm()
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("RefreshOpenidService - fail parse form")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusBadRequest))
		return
	}

	r.Form.Add("role", role)
	r.Form.Add("path", "refreshopenid") // will inform the CustomizeHandler
	r.Form.Add("sub", emailOrAPIsvcID.(string))

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("RefreshOpenidService - exit successfully")

	// if error during refreshing, return err
	err = t.srv.RefreshOpenidToken(context.Background(), w, r, data)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("RefreshOpenidService - t.srv.RefreshOpenidToken execution failed")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden, "/refreshopenid", err.Error()))
	}
}

// TokenService handle the request to provide the jwt(2nd request of the oauth/openid protocol)
func (t *TokenService) TokenService(w http.ResponseWriter, r *http.Request, authHeader string) e.IError {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("TokenService - hit service", r)

	err := t.srv.HandleTokenRequest(w, r)
	if err != nil {
		return e.NewCustomHTTPStatus(e.StatusInternalServerError)
	}
	return nil
}

// JwtGetdataService valid the jwt and return the data it holds
func (t *TokenService) JwtGetdataService(w http.ResponseWriter, r *http.Request, cookie *http.Cookie) (map[string]interface{}, e.IError) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("JwtGetdataService - hit service", r)

	jwt := cookie.Value
	// fmt.Println("see the cookie value: ", jwt)

	// validate the token
	keyID := "theKeyID"
	secretKey := "mySecretKey"
	encoding := "HS256"

	data, err := t.srv.HandleJWTokenGetdata(context.TODO(), r, jwt, keyID, secretKey, encoding)
	if err != nil {

		switch err.Error() {
		case "expired jwt token":
			return nil, e.NewCustomHTTPStatus(e.StatusUnauthorized)
		case "invalid jwt token":
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg("JwtGetdataService - t.srv.HandleJWTokenGetdata execution failed - invalid jwt token")
			return nil, e.NewCustomHTTPStatus(e.StatusForbidden)
		default:
			// TODO what about the default ??
			return nil, e.NewCustomHTTPStatus(e.StatusForbidden)
		}
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("JwtGetdataService - exit successfully")

	return data, nil
}

// JwtValidationService validate the jwt(here optional, as it's handle by the broker)
func (t *TokenService) JwtValidationService(r *http.Request, cookie *http.Cookie) e.IError {

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("JwtValidationService - hit service", r)

	jwt := cookie.Value

	// validate the token
	keyID := "theKeyID"
	secretKey := "mySecretKey"
	encoding := "HS256"

	err := t.srv.HandleJWTokenValidation(context.TODO(), r, jwt, keyID, secretKey, encoding)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("JwtValidationService - t.srv.HandleJWTokenGetdata execution failed - invalid jwt token")
		return e.NewCustomHTTPStatus(e.StatusUnauthorized)
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msg("JwtValidationService - exit successfully")

	// token is valid, return nil
	return nil
}

// Endpoint to validate token and permission
// func (t *TokenService) ValidPermissionService(w http.ResponseWriter, r *http.Request, permission string, token *http.Cookie) map[string]interface{} {
func (t *TokenService) ValidPermissionService(w http.ResponseWriter, r *http.Request) (map[string]interface{}, e.IError) {

	var ce e.IError
	// validate the token
	token, err := t.srv.ValidationBearerToken(r)
	if err != nil {
		// w.WriteHeader(http.StatusBadRequest)
		ce = e.NewCustomHTTPStatus(e.StatusBadRequest)
		// http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, ce
	}

	permission := r.URL.Query().Get("permission")

	// validate the permission
	switch permission {
	case "read":
		log.Println("In read permission")
		if !strings.Contains(token.GetScope(), "read") && !strings.Contains(token.GetScope(), "all") {
			// http.Error(w, "Unauthorized", http.StatusBadRequest)
			return nil, e.NewCustomHTTPStatus(e.StatusBadRequest)
		}

	case "write":
		log.Println("In write permission")
		if !strings.Contains(token.GetScope(), "write") && !strings.Contains(token.GetScope(), "all") {
			fmt.Println("do not have Write permission.")
			// http.Error(w, "Unauthorized", http.StatusBadRequest)
			return nil, e.NewCustomHTTPStatus(e.StatusBadRequest)
		}

	case "all":
		log.Println("In all permission")
		if !strings.Contains(token.GetScope(), "all") {
			fmt.Println("do not have All permission.")
			// http.Error(w, "Unauthorized", http.StatusBadRequest)
			return nil, e.NewCustomHTTPStatus(e.StatusBadRequest)
		}
	default:
		log.Println("In default permission")
		// http.Error(w, "Unauthorized", http.StatusBadRequest)
		return nil, e.NewCustomHTTPStatus(e.StatusBadRequest)
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"client_id":  token.GetClientID(),
		"user_id":    token.GetUserID(),
		"permission": token.GetScope(),
	}

	return data, nil
}
