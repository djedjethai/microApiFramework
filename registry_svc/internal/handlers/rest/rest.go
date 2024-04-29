package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/config"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/services"
	"net/http"
)

var restConfigs config.IRestConfig

func Handler(configs config.IRestConfig, services *services.JWTokenService) *mux.Router {
	router := mux.NewRouter()

	// set the configs as global in the package
	restConfigs = configs

	// handle all grpc schemas download
	grpcRouter := router.PathPrefix("/grpc/").Subrouter()
	NewGrpcHandler(grpcRouter).RunGrpcRest()

	// handle all configs(like certificates) download
	configsRouter := router.PathPrefix("/configs/").Subrouter()
	NewConfigsHandler(configsRouter).RunConfigsRest()

	// healthcheck endpoint
	router.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("registry_svc, reach the healthcheck..................")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// }).Name("healthcheck") // do nothing

	// middleware
	// tokenService := services.NewTokenService() // is not a pointer
	am := NewAuthMiddleware(services)
	router.Use(am.authorizationHandler())

	return router
}

// serveFilesHandler serve all handlers
func serveFilesHandler(storageDir string, files ...string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		v := vars["version"]
		p := vars["package"]
		t := vars["type"]
		rd := vars["repodir"]

		ah, err := NewAgentsHandler(w, files, storageDir, v, p, t, rd)
		if err != nil {
			http.Error(w, err.Error(), err.GetCode())
			return
		}

		err = ah.run()
		if err != nil {
			w.WriteHeader(err.GetCode())
			fmt.Fprint(w, err.Error())
			return
		}
	}
}
