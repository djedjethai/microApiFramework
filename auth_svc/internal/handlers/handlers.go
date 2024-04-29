package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type Handlers struct {
	authHandler  IAuthenticationHandler
	tokenHandler ITokenHandler
}

func NewHandlers(a IAuthenticationHandler, t ITokenHandler) Handlers {
	return Handlers{
		authHandler:  a,
		tokenHandler: t,
	}
}

func (h Handlers) Run(portvar int) {
	// TODO set all handler in a separate func
	// Endpoints for the front-end
	// (use this service for the example but a specific users' service may be better)
	http.HandleFunc("/v1/signup", h.authHandler.Signup)
	http.HandleFunc("/v1/signin", h.authHandler.Signin)
	http.HandleFunc("/v1/signout", h.authHandler.Signout)

	// Endpoints for the backend services to authenticate and get their token
	http.HandleFunc("/v1/apiauth", h.authHandler.ApiAuth)
	http.HandleFunc("/v1/refreshopenid", h.tokenHandler.RefreshOpenid)

	// Endpoints specific to validate the authorization
	// http.HandleFunc("/v1/oauth/authorize", h.authHandler.Authorize)
	http.HandleFunc("/v1/oauth/token", h.tokenHandler.Token)

	// Endpoint which validate a client's token and the given permission
	http.HandleFunc("/v1/jwtvalidation", h.tokenHandler.JwtValidation)
	http.HandleFunc("/v1/jwtgetdata", h.tokenHandler.JwtGetdata)
	http.HandleFunc("/v1/permission", h.tokenHandler.ValidPermission)

	// healthcheck endpoint
	http.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("auth-svc, reach the healthcheck..................")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server is running at %d port.\n", portvar)
	log.Printf("Point your OAuth client Auth endpoint to %s:%d%s", "http://localhost", portvar, "/oauth/authorize")
	log.Printf("Point your OAuth client Token endpoint to %s:%d%s", "http://localhost", portvar, "/oauth/token")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portvar), nil))

}
