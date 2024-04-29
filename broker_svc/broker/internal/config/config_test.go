package config

import (
	"fmt"
	"gitlab.com/grpasr/common/tests"
	"os"
	"testing"
)

func Test_default_configs(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	os.Setenv("GOENV", "localhost")
	os.Setenv("CONFIG_FILE_PATH", "../../configs/v1/servicesAddr/")

	os.Setenv("PATH_TO_TLS", "../../configs/v1/certificates/")

	conf, _ := SetConfigs()

	tests.MaybeFail("Test_default_configs",
		tests.Expect(fmt.Sprintf("%v", conf.global), "&{localhost brokerSvc }"),
		tests.Expect(fmt.Sprintf("%v", conf.http), "&{localhost 8080}"),
		tests.Expect(fmt.Sprintf("%v", conf.grpc), "&{localhost 50002}"),
		tests.Expect(fmt.Sprintf("%v", conf.jwtRequestConfig), "&{http://localhost:9096/v1 apiauth http://localhost:9096/v1/oauth/token exampleCodeVerifier brokerSvc brokerSvcSecret read, openid}"),
		tests.Expect(fmt.Sprintf("%v", conf.SVCSGetServices()), "map[order:{localhost 50001} preorder:{localhost 50001}]"),
		tests.Expect(fmt.Sprintf("%v", conf.OBSGetSampling()), "0.6"),
		tests.Expect(fmt.Sprintf("%v", conf.OBSGetScratchDelay()), "30"),
		tests.Expect(fmt.Sprintf("%v", conf.OBSGetCollectorEndpoint()), "otel_collector:4317"),
		tests.Expect(conf.ClientTLSConfig != nil, true),
		tests.Expect(conf.ServerTLSConfig != nil, true),
		tests.Expect(fmt.Sprintf("%v", conf.TOKENGetSignedKey()), "mySecretKey"),
	)
}

func Test_configs_development_env(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// set global
	os.Setenv("GOENV", "development")
	os.Setenv("SERVICE_NAME", "serviceName")
	os.Setenv("CONFIG_FILE_PATH", "../../configs/v1/servicesAddr/")

	// set http
	os.Setenv("SERVICE_ADDRESS", "serviceAddress")
	os.Setenv("SERVICE_PORT", "servicePort")

	// set Grpc
	os.Setenv("GRPC_BROKER_SVC_PORT", "grpcBrokerSvcPort")

	// set JWTRequestConfig
	os.Setenv("SERVICE_URL", "serviceUrl")
	os.Setenv("AUTH_SVC_URL", "authSvcUrl")
	os.Setenv("AUTH_SVC_PATH", "authSvcPath")
	os.Setenv("AUTH_SVC_TOKEN_ENDPOINT", "authSvcTokenEndpoint")
	os.Setenv("CODE_VERIFIER", "codeVerifier")
	os.Setenv("SERVICE_KEY_ID", "serviceKeyID")
	os.Setenv("SERVICE_SECRET_KEY", "serviceSecretKey")
	os.Setenv("SCOPE", "scope")

	// set jwt token signedKey
	os.Setenv("SIGNED_KEY", "newSecret")

	// set observability
	os.Setenv("OBS_SAMPLING", "1")
	os.Setenv("OBS_SCRATCH_DELAY", "2")
	os.Setenv("OBS_COLLECTOR_ENDPOINT", "collector")

	conf, _ := SetConfigs()

	tests.MaybeFail("Test_default_configs",
		// tests.Expect(fmt.Sprintf("%v", conf.Global), "&{golangEnv serviceName serviceUrl}"),
		tests.Expect(fmt.Sprintf("%v", conf.global), "&{development serviceName }"),
		tests.Expect(fmt.Sprintf("%v", conf.http), "&{serviceAddress servicePort}"),
		tests.Expect(fmt.Sprintf("%v", conf.grpc), "&{localhost 50002}"),
		tests.Expect(fmt.Sprintf("%v", conf.jwtRequestConfig), "&{authSvcUrl authSvcPath authSvcTokenEndpoint codeVerifier serviceKeyID serviceSecretKey scope}"),
		tests.Expect(fmt.Sprintf("%v", conf.SVCSGetServices()), "map[order:{order 50001} preorder:{preOrder 50001}]"),
		tests.Expect(fmt.Sprintf("%v", conf.SVCSGetServices()["order"]), "{order 50001}"),
		tests.Expect(fmt.Sprintf("%v", conf.OBSGetSampling()), "1"),
		tests.Expect(fmt.Sprintf("%v", conf.OBSGetScratchDelay()), "2"),
		tests.Expect(fmt.Sprintf("%v", conf.OBSGetCollectorEndpoint()), "collector"),
		tests.Expect(fmt.Sprintf("%v", conf.TOKENGetSignedKey()), "newSecret"),
	)
}
