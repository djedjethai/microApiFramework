package main

import (
	"flag"
	"fmt"
	// "fmt"
	"log"
	// "os"

	"gitlab.com/grpasr/asonrythme/auth_svc/internal/config"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/handlers"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/repository"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/services"

	"github.com/djedjethai/go-oauth2-openid/errors"
	"github.com/djedjethai/go-oauth2-openid/manage"
	"github.com/djedjethai/go-oauth2-openid/server"
	obs "gitlab.com/grpasr/common/observability"

	mongo "github.com/djedjethai/mongo-openid"
)

var (
	// dumpvar bool
	// idvar     string
	// secretvar string
	// domainvar string
	portvar int
)

// TODO
// right now as a client is register in db, it won't be updated in case of modification
// means the updates won't take effect........ see in the mongo client...

func init() {
	// TODO change all of that using a YAML file
	// credential for the client
	// flag.BoolVar(&dumpvar, "d", true, "Dump requests and responses")
	// flag.StringVar(&idvar, "i", "222222", "The client id being passed in")
	// flag.StringVar(&secretvar, "s", "22222222", "The client secret being passed in")
	// flag.StringVar(&domainvar, "r", "http://localhost:80", "The domain of the redirect url")
	flag.IntVar(&portvar, "p", 9096, "the base port for the server")

}

func main() {
	flag.Parse()

	conf, err := setConfigs()
	if err != nil {
		log.Fatal("setConfigs failed: ", err)
	}

	fmt.Println("see conf: ", conf)

	// set the observability
	obs.SetObservabilityFacade(conf.GlbGetServiceName())
	// TODO create an endpoint to switch between dev and prod
	obs.Logging.SetLoggingEnvToDevelopment()

	// note that, the accessToken(and the jwt_access_token by cons) for APIServer role
	// are set to 6 month, to avoid(for the moment)
	// implementing the refresh token logic on the servers
	mc := manage.ManagerConfig{
		AuthorizeCodeTokenCfgAccess:      2,
		AuthorizeCodeTokenCfgRefresh:     24 * 15,
		AuthorizeCodeAPIServerCfgAccess:  24 * 180,
		AuthorizeCodeAPIServerCfgRefresh: 24 * 360,
	}

	manager := manage.NewDefaultManager(mc)

	// set connectionTimeout(7s) and the requestsTimeout(5s) // is optional
	storeConfigs := mongo.NewStoreConfig(7, 7)

	mongoConf := createMongoAuthConfig(conf)

	fmt.Println("see mongo conf: ", mongoConf)

	// use mongodb token store
	manager.MapTokenStorage(
		mongo.NewTokenStore(mongoConf, storeConfigs), // with timeout
		// mongo.NewTokenStore(mongoConf), // no timeout
	)

	clientStore := mongo.NewClientStore(mongoConf, storeConfigs) // with timeout
	// clientStore := mongo.NewClientStore(mongoConf) // no timeout

	manager.MapClientStorage(clientStore)

	// register all app services
	registerServices(clientStore, conf)

	srv := server.NewServer(server.NewConfig(), manager)

	// set the oauth package to work without browser
	// the token will be return as a json payload
	srv.SetModeAPI()

	// set the service domain
	repos := repository.NewRepository(conf)

	// set the service services
	authService := services.NewAuthenticationService(srv, repos)
	oauth2Service := services.NewOauth2Service(srv, repos)
	tokenService := services.NewTokenService(srv, repos)

	// handlers will handle all handlers
	authHandler := handlers.NewAuthenticationHandler(authService)
	tokenHandler := handlers.NewTokenHandler(tokenService)
	// handler := handlers.NewHandlers(dumpvar, srv, repos)
	handlersHandle := handlers.NewHandlers(authHandler, tokenHandler)

	// set the authorization staff
	srv.SetUserAuthorizationHandler(oauth2Service.UserAuthorizeService)
	// set the openid staff
	srv.SetUserOpenidHandler(oauth2Service.UserOpenidService)
	// set the func to, before to send it, customize the token payload
	srv.SetCustomizeTokenPayloadHandler(oauth2Service.UserCustomizeTokenPayloadService)

	// TODO add our logs
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	// TODO add our logs
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	handlersHandle.Run(portvar)

}

func createMongoAuthConfig(conf *config.Config) *mongo.Config {

	var mongoConf *mongo.Config
	if conf.MgoGetCluster() {
		// "mongodb://localhost:27017,localhost:28017,localhost:29017/?replicaSet=myReplicaSet",
		dsn := fmt.Sprintf("%v/?replicaSet=%s", conf.MgoGetURL(), conf.MgoGetReplicaSetName())
		mongoConf = mongo.NewConfigReplicaSet(
			dsn,
			conf.MgoGetAuthDatabaseName(),
		)
	} else {
		mongoConf = mongo.NewConfigNonReplicaSet(
			conf.MgoGetURL(),
			conf.MgoGetAuthDatabaseName(),
			conf.MgoGetUsername(),
			conf.MgoGetPassword(),
			conf.GlbGetServiceName())
	}

	return mongoConf
}
