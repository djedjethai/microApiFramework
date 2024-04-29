package loader

import (
	"fmt"
	"gitlab.com/grpasr/common/tests"
	"os"
	"testing"
)

func Test_default_configs(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	conf := setConfigs()

	tests.MaybeFail("Test_default_configs",
		tests.Expect(fmt.Sprintf("%v", conf.Global), "&{localhost brokerSvc http://localhost:8080}"),
		tests.Expect(fmt.Sprintf("%v", conf.JWTRequestConfig), "&{http://localhost:9096/v1 apiauth http://localhost:9096/v1/oauth/token exampleCodeVerifier brokerSvc brokerSvcSecret read, openid}"),
	)
}

func Test_configs_with_varenv(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// Set environment variable
	os.Setenv("GOENV", "golangEnv")
	os.Setenv("SERVICE_NAME", "serviceName")
	os.Setenv("SERVICE_URL", "serviceUrl")
	os.Setenv("AUTH_SVC_URL", "authSvcUrl")
	os.Setenv("AUTH_SVC_PATH", "authSvcPath")
	os.Setenv("AUTH_SVC_TOKEN_ENDPOINT", "authSvcTokenEndpoint")
	os.Setenv("CODE_VERIFIER", "codeVerifier")
	os.Setenv("SERVICE_KEY_ID", "serviceKeyID")
	os.Setenv("SERVICE_SECRET_KEY", "serviceSecretKey")
	os.Setenv("SCOPE", "scope")

	conf := setConfigs()

	tests.MaybeFail("Test_default_configs",
		tests.Expect(fmt.Sprintf("%v", conf.Global), "&{golangEnv serviceName serviceUrl}"),
		tests.Expect(fmt.Sprintf("%v", conf.JWTRequestConfig), "&{authSvcUrl authSvcPath authSvcTokenEndpoint codeVerifier serviceKeyID serviceSecretKey scope}"),
	)
}
