package handlers

import (
	"encoding/json"
	"net/http"

	"gitlab.com/grpasr/asonrythme/auth_svc/internal/services"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
)

// const (
// 	authServerURL string = "http://localhost:9096"
// )

type ITokenHandler interface {
	RefreshOpenid(w http.ResponseWriter, r *http.Request)
	Token(w http.ResponseWriter, r *http.Request)
	JwtGetdata(w http.ResponseWriter, r *http.Request)
	JwtValidation(w http.ResponseWriter, r *http.Request)
	ValidPermission(w http.ResponseWriter, r *http.Request)
}

type TokenHandler struct {
	srv services.ITokenService
}

func NewTokenHandler(srv services.ITokenService) ITokenHandler {
	return TokenHandler{srv}
}

func (t TokenHandler) RefreshOpenid(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("RefreshOpenid - hit handler", r)

	// w.Header().Set("Content-Type", "application/json")

	// fmt.Println("Into the RefreshOpenid, get email from expiredToken: ")

	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(e.NewCustomHTTPStatus(e.StatusForbidden))
		return
	}

	t.srv.RefreshOpenidService(w, r, cookie)

}

func (t TokenHandler) Token(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("Token - hit handler", r)

	authHeader := r.Header.Get("Authorization")

	ce := t.srv.TokenService(w, r, authHeader)
	if ce != nil {
		switch ce.GetCode() {
		case http.StatusUnauthorized:
			w.WriteHeader(http.StatusUnauthorized)
		case http.StatusInternalServerError:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			ce = e.NewCustomHTTPStatus(e.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ce)
	}
}

func (t TokenHandler) JwtGetdata(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("JwtGetdata - hit handler", r)

	var ce e.IError
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		ce = e.NewCustomHTTPStatus(e.StatusForbidden)
	} else {
		data, ce := t.srv.JwtGetdataService(w, r, cookie)
		if ce != nil {
			switch ce.GetCode() {
			case http.StatusUnauthorized:
				w.WriteHeader(http.StatusUnauthorized)
			case http.StatusOK:
				w.WriteHeader(http.StatusOK)
			default:
				ce = e.NewCustomHTTPStatus(e.StatusInternalServerError)
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
			return
		}
	}
	json.NewEncoder(w).Encode(ce)
}

// TODO this have to be handle by the apigateway, here for testing
func (t TokenHandler) JwtValidation(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("JwtValidation - hit handler", r)

	var ce e.IError
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		ce = e.NewCustomHTTPStatus(e.StatusForbidden)
		json.NewEncoder(w).Encode(ce)
		return

	}

	ce = t.srv.JwtValidationService(r, cookie)
	if ce != nil {
		switch ce.GetCode() {
		case http.StatusUnauthorized:
			w.WriteHeader(http.StatusUnauthorized)
		default:
			ce = e.NewCustomHTTPStatus(e.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ce)
	} else {
		// err is nil so token is valid, Respond with an OK status (200)
		w.WriteHeader(http.StatusOK)
	}

}

// Endpoint to validate token and permission
func (t TokenHandler) ValidPermission(w http.ResponseWriter, r *http.Request) {
	obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
		Msgf("ValidPermission - hit handler", r)

	data, ce := t.srv.ValidPermissionService(w, r)
	if ce != nil {
		switch ce.GetCode() {
		case http.StatusBadRequest:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			ce = e.NewCustomHTTPStatus(e.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ce)
		return
	}

	ej := json.NewEncoder(w)
	ej.SetIndent("", "  ")
	ej.Encode(data)
}

// func dumpRequest(writer io.Writer, header string, r *http.Request) error {
// 	data, err := httputil.DumpRequest(r, true)
// 	if err != nil {
// 		return err
// 	}
// 	writer.Write([]byte("\n" + header + ": \n"))
// 	writer.Write(data)
//
// 	// TODO see where are the datas
// 	log.Println("enddump")
// 	return nil
// }
