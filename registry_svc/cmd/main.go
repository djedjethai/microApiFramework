package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/config"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/handlers/rest"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/services"
	"gitlab.com/grpasr/common/apiserver"
	"log"
	"net/http"
	// "os"
)

var (
	jwtoken string
	conf    *config.Config
)

func init() {
	config.SetVarEnv()

	conf = config.NewConfig()

	// get jwtoken from auth_svc
	apiServerAuth := apiserver.NewAPIserverAuth(
		conf.JwtGetAuthSvcURL(),           // "http://localhost:9096/v1", // all varEnv
		conf.JwtGetAuthSvcPath(),          // "apiauth",
		conf.HTTPGetHTTPFormatedURL(),     // "http://localhost:4000",
		conf.JwtGetAuthSvcTokenEndpoint(), // "http://localhost:9096/v1/oauth/token",
		conf.JwtGetCodeVerifier(),         //"exampleCodeVerifier",
		conf.JwtGetServiceKeyID(),         // "registrySvc",
		conf.JwtGetServiceSecretKey(),     // "registrySvcSecret",
		conf.JwtGetScope(),                // "read, openid",
	)

	err := apiServerAuth.Run(context.TODO(), int8(3), int8(5))
	if err != nil {
		log.Fatal("registry_svc error authentication: ", err)
	}

	jwtoken = apiServerAuth.GetToken()

	fmt.Println("see the registry_svc token: ", jwtoken)
}

func main() {
	restConfigs := config.NewRestConfig()

	svc := services.NewServices()

	router := rest.Handler(restConfigs, svc.JWTokenService)

	startServer(router, conf)
}

func startServer(router *mux.Router, configs *config.Config) {
	// starting server
	// address := os.Getenv("SERVER_ADDRESS")
	// address := configs.HTTPGetAddress()
	// port := os.Getenv("SERVER_PORT")
	port := configs.HTTPGetPort()
	fmt.Printf("registry_svc listen on port: :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
