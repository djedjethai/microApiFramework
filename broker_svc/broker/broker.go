package broker

import (
	"context"
	// "fmt"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/grpc/clients"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/grpc/servers"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/rest"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services"
	"gitlab.com/grpasr/common/apiserver"
	obs "gitlab.com/grpasr/common/observability"
	"log"
)

// see if a port is in use: sudo lsof -i :50002
// kill the process: sudo kill xxxx
// const (
// 	samplingRatio     float64 = 0.6
// 	scratchDelay      int     = 30
// 	collectorEndpoint         = "localhost:4317"
// )

var conf *config.Config

func init() {

	var err error
	conf, err = setConfigs()
	if err != nil {
		log.Fatal("Error setting configs: ", err)
	}

	// get jwtoken from auth_svc
	apiServerAuth := apiserver.NewAPIserverAuth(
		conf.JwtGetAuthSvcURL(),           // "http://localhost:9096/v1", // all varEnv
		conf.JwtGetAuthSvcPath(),          // "apiauth",
		conf.HTTPGetHTTPFormatedURL(),     // "http://localhost:8080",
		conf.JwtGetAuthSvcTokenEndpoint(), // "http://localhost:9096/v1/oauth/token",
		conf.JwtGetCodeVerifier(),         //"exampleCodeVerifier",
		conf.JwtGetServiceKeyID(),         // "brokerSvc",
		conf.JwtGetServiceSecretKey(),     // "brokerSvcSecret",
		conf.JwtGetScope(),                // "read, openid",
	)

	err = apiServerAuth.Run(context.TODO(), int8(3), int8(5))
	if err != nil {
		log.Fatal("Order svc error authentication: ", err)
	}

	conf.GlbSetJWToken(apiServerAuth.GetToken())
	log.Println("broker_svc/broker see the jwtoken: ", conf.GlbGetJWToken())
}

func Run() {
	// set observability
	obs.SetObservabilityFacade(conf.GlbGetServiceName())
	// TODO create an endpoint to switch between dev and prod
	obs.Logging.SetLoggingEnvToDevelopment()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if conf.GlbGetenv() != "localhost" {
		tp, err := obs.Tracing.SetupTracing(
			ctx,
			conf.ClientTLSConfig,
			conf.OBSGetSampling(),
			conf.GlbGetServiceName(),
			conf.OBSGetCollectorEndpoint(),
			conf.GlbGetenv())
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHFatal(), ctx).
				Err(err).
				Send()
		}
		defer tp.Shutdown(ctx)

		mp, err := obs.Metrics.SetupMetrics(
			ctx,
			conf.ClientTLSConfig,
			conf.OBSGetScratchDelay(),
			conf.GlbGetServiceName(),
			conf.OBSGetCollectorEndpoint(),
			conf.GlbGetenv())
		if err != nil {
			obs.Logging.NewLogHandler(obs.Logging.LLHFatal(), ctx).
				Err(err).
				Send()
		}
		defer mp.Shutdown(ctx)

	}

	grpcClients, err := clients.NewGRPCClients(conf)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHFatal(), ctx).
			Err(err).
			Send()
	}

	restServices := services.NewRestServices(grpcClients, conf)
	grpcServices := services.NewGrpcServices(grpcClients, conf)

	// run grpcServer
	grpcServer, err := servers.NewGRPCServers(grpcServices, conf)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHFatal(), ctx).
			Err(err).
			Send()
	}
	grpcServer.GRPCServersListen()

	// run restServer
	rest.Handler(restServices, conf)
}
