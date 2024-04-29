package config

import (
	"github.com/spf13/viper"
	"gitlab.com/grpasr/common/tests"
	"os"
	"testing"
)

func Test_appConfig_get_default_configs(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	SetVarEnv()
	defer viper.Reset()

	c := NewConfig()
	rc := NewRestConfig()

	tests.MaybeFail("createNewCodeError",
		tests.Expect(c.Global.golangEnv, "localhost"),
		tests.Expect(c.Global.serviceName, "registry-svc"),
		tests.Expect(c.Http.address, "127.0.0.1"),
		tests.Expect(c.Http.port, "4000"),
		tests.Expect(rc.pathToStorage, ".."),
		tests.Expect(rc.grpcDirectory, "api"),
		tests.Expect(rc.configsDirectory, "configs"),
	)
}

func Test_appConfig_test_setters_and_guetters(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	SetVarEnv()
	defer viper.Reset()

	c := NewConfig()

	rc := NewRestConfig()
	rc.RESTSetPathToStorage("../new")
	rc.RESTSetGrpcDirectory("newGrpc")
	rc.RESTSetConfigsDirectory("newConfigs")

	tests.MaybeFail("createNewCodeError",
		tests.Expect(c.GLBGetenv(), "localhost"),
		tests.Expect(c.GLBGetServiceName(), "registry-svc"),
		tests.Expect(c.HTTPGetAddress(), "127.0.0.1"),
		tests.Expect(c.HTTPGetPort(), "4000"),
		tests.Expect(rc.RESTGetPathToStorage(), "../new"),
		tests.Expect(rc.RESTGetGrpcDirectory(), "newGrpc"),
		tests.Expect(rc.RESTGetConfigsDirectory(), "newConfigs"),
	)
}

// Keep this test the last one as the os.varEnv persist
func Test_appConfig_default_configs_are_overwritten(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// Set some manually for testing
	os.Setenv("GO_ENV", "test_env")
	os.Setenv("SERVICE_NAME", "test-registry-svc")
	os.Setenv("HTTP_ADDRESS", "123.3.4.5")
	os.Setenv("HTTP_PORT", "5000")
	os.Setenv("PATH_STORAGE", "../path_env")
	os.Setenv("GRPC_DIRECTORY", "test_api")
	os.Setenv("CONFIGS_DIRECTORY", "test_configs")

	SetVarEnv()
	defer viper.Reset()

	c := NewConfig()
	rc := NewRestConfig()

	tests.MaybeFail("createNewCodeError",
		tests.Expect(c.Global.golangEnv, "test_env"),
		tests.Expect(c.Global.serviceName, "test-registry-svc"),
		tests.Expect(c.Http.address, "123.3.4.5"),
		tests.Expect(c.Http.port, "5000"),
		tests.Expect(rc.pathToStorage, "../path_env"),
		tests.Expect(rc.grpcDirectory, "test_api"),
		tests.Expect(rc.configsDirectory, "test_configs"),
	)
}
