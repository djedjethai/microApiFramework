package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/rest/orderRest"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services"
	obs "gitlab.com/grpasr/common/observability"
	"net/http"
)

func Handler(services services.RestServices, configs *config.Config) {

	router := mux.NewRouter()

	// Order service routes
	orderRouter := router.PathPrefix("/order").Subrouter()
	orderRest.NewOrderRest(orderRouter, services.OrderService).RunOrderRest()

	// middleware
	// tokenService := services.NewTokenService() // is not a pointer
	am := authMiddleware{services.JWTokenService}
	router.Use(am.authorizationHandler())

	// healthcheck endpoint
	router.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
			Msg("reach the healthcheck")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// logging endpoint
	router.HandleFunc("/v1/logging", func(w http.ResponseWriter, r *http.Request) {
		obs.Logging.NewLogHandler(obs.Logging.LLHDebug()).
			Msg("reach the logging endpoint")
		logEnv := obs.Logging.GetLoggingEnv()
		switch logEnv {
		case "production":
			obs.Logging.SetLoggingEnvToDevelopment()
		case "development":
			obs.Logging.SetLoggingEnvToProduction()
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// starting server
	port := configs.HTTPGetPort()
	obs.Logging.NewLogHandler(obs.Logging.LLHInfo()).
		Str("HTTP listen on port: ", port).
		Send()
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHFatal()).
			Err(err).
			Send()
	}
}
