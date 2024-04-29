package rest

import (
	"context"
	"encoding/json"
	"fmt"
	// "github.com/gorilla/mux"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/services"
	e "gitlab.com/grpasr/common/errors/json"
	"net/http"
	"strings"
)

type authMiddleware struct {
	tokenService *services.JWTokenService
}

func NewAuthMiddleware(jwtSvc *services.JWTokenService) authMiddleware {
	return authMiddleware{jwtSvc}
}

func (a authMiddleware) authorizationHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Hit the authMiddleware......: ", r.URL.Path)
			authHeader := r.Header.Get("Authorization")
			// fmt.Println("see the authorization: ", authHeader)

			// Retrieve route variables using mux.Vars
			//vars := mux.Vars(r) // but return nothing...
			// _, ok := vars["healthcheck"]

			// do not apply the token validation for the healthcheck
			if r.URL.Path != "/v1/health" {

				token := getTokenFromHeader(authHeader)
				if token == "" {
					w.WriteHeader(http.StatusForbidden)
					w.Header().Set("Content-Type", "application/json")
					ce := e.NewCustomHTTPStatus(e.StatusForbidden, "", "jwtoken not found")
					json.NewEncoder(w).Encode(ce)
					return
				}

				// validate the token
				_, err := a.tokenService.JWTokenIsValidToken(context.TODO(), token)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					switch err.GetCode() {
					case 401:
						w.WriteHeader(http.StatusUnauthorized)
					case 403:
						w.WriteHeader(http.StatusForbidden)
					}
					json.NewEncoder(w).Encode(err)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getTokenFromHeader(header string) string {
	/*
	   token is coming in the format as below
	   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50cyI6W.yI5NTQ3MCIsIjk1NDcyIiw"
	*/
	splitToken := strings.Split(header, "Bearer")
	if len(splitToken) == 2 {
		return strings.TrimSpace(splitToken[1])
	}

	return ""
}
