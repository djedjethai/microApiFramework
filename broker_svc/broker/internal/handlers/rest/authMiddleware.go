package rest

import (
	"fmt"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/jwtokenService"
	// "github.com/gorilla/mux"
	"net/http"
	"strings"
)

type authMiddleware struct {
	tokenService *jwtokenService.JWTokenService
}

func (a authMiddleware) authorizationHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Hit the authMiddleware")
			next.ServeHTTP(w, r)
			// currentRoute := mux.CurrentRoute(r)
			// currentRouteVars := mux.Vars(r)
			// authHeader := r.Header.Get("Authorization")

			// logger.Info("see the vars from route: " + fmt.Sprintf("%v", currentRouteVars))

			// if authHeader != "" {
			// 	token := getTokenFromHeader(authHeader)

			// 	logger.Info("token in account route: " + token)

			// 	// TODO set a whiteLists of allowed routes
			// 	// if ok = whiteList[currentRoute.GetName()]{ no check :) }
			// 	isAuthorized := a.tokenService.IsAuthorized(token, currentRoute.GetName(), currentRouteVars)

			// 	if isAuthorized {
			// 		next.ServeHTTP(w, r)
			// 	} else {
			// 		appErr := errs.AppError{
			// 			Code:    http.StatusForbidden,
			// 			Message: "Unauthorized"}
			// 		writeResponse(w, appErr.Code, appErr.AsMessage())
			// 	}
			// } else {
			// 	writeResponse(w, http.StatusUnauthorized, "missing token")
			// }
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
