package config

import (
	// "fmt"
	"gitlab.com/grpasr/common/tests"
	"os"
	"testing"
)

const (
	serviceName            = "authA"
	mongoAuthDatabaseName  = "oauth2A"
	mongoUsersDatabaseName = "authenticationA"
	mongoReplicaSetName    = "myReplicaSetA"
	mongoUsername          = "adminA"
	mongoPassword          = "passwordA"
	isMongoCluster         = "true"
	jwtAlgorithm           = "HS256A"
	jwtUserKeyID           = "userKeyIDDefaultA"
	jwtUserSecretkey       = "userSecretkeyDefaultA"
	jwtServiceKeyID        = "serviceKeyIDDefaultA"
	jwtServiceSecretkey    = "serviceSecretkeyDefaultA"
	redisPort              = "63790"
	redisAddress           = "redisA"
	redisMaxIdle           = "809"
	redisMaxActive         = "120009"
)

func Test_default_configs(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	os.Setenv("GOENV", "localhost")
	os.Setenv("CONFIG_FILE_PATH", "../../configs/v1/")
	os.Setenv("CONFIG_FILE_NAME", "servicesTest")

	conf, _ := SetConfigs()

	svcs := conf.SVCGetServices()

	tests.MaybeFail("Test_default_configs",
		tests.Expect(conf.GlbGetenv(), "localhost"),
		tests.Expect(conf.GlbGetServiceName(), "auth"),
		tests.Expect(svcs["frontend"].SVCGetID(), "222222"),
		tests.Expect(svcs["frontend"].SVCGetSecret(), "22222222"),
		tests.Expect(svcs["frontend"].SVCGetDomain(), "http://localhost:80"),
		tests.Expect(svcs["broker_svc"].SVCGetID(), "brokerSvc"),
		tests.Expect(svcs["registry_svc"].SVCGetDomain(), "http://localhost:4000"),
		tests.Expect(conf.MgoGetAuthDatabaseName(), mongoAuthDatabaseNameDefault),
		tests.Expect(conf.MgoGetUsersDatabaseName(), mongoUsersDatabaseNameDefault),
		tests.Expect(conf.MgoGetReplicaSetName(), mongoReplicaSetNameDefault),
		tests.Expect(conf.MgoGetCluster(), isMongoClusterDefault),
		tests.Expect(conf.MgoGetURL(), mongoURLDefault),
		tests.Expect(conf.MgoGetUsername(), mongoUsernameDefault),
		tests.Expect(conf.MgoGetPassword(), mongoPasswordDefault),
		tests.Expect(conf.RdsGetPort(), redisPortDefault),
		tests.Expect(conf.RdsGetAddress(), redisAddressDefault),
		tests.Expect(conf.RdsGetMaxIdle(), redisMaxIdleDefault),
		tests.Expect(conf.RdsGetMaxActive(), redisMaxActiveDefault),
		tests.Expect(conf.JwtGetAlgorithm(), jwtAlgorithmDefault),
		tests.Expect(conf.JwtGetUserKeyID(), jwtUserKeyIDDefault),
		tests.Expect(conf.JwtGetUserSecretkey(), jwtUserSecretkeyDefault),
		tests.Expect(conf.JwtGetServiceKeyID(), jwtServiceKeyIDDefault),
		tests.Expect(conf.JwtGetServiceSecretkey(), jwtServiceSecretkeyDefault),
	)
}

func Test_varenv_configs(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	os.Setenv("GOENV", "development")
	os.Setenv("SERVICE_NAME", serviceName)
	os.Setenv("CONFIG_FILE_PATH", "../../configs/v1/")
	os.Setenv("CONFIG_FILE_NAME", "servicesTest")
	os.Setenv("MONGO_USERNAME", mongoUsername)
	os.Setenv("MONGO_PASSWORD", mongoPassword)
	os.Setenv("MONGO_CLUSTER", isMongoCluster)
	os.Setenv("MONGO_AUTH_DATABASE", mongoAuthDatabaseName)
	os.Setenv("MONGO_USERS_DATABASE", mongoUsersDatabaseName)
	os.Setenv("MONGO_REPLICASET_NAME", mongoReplicaSetName)
	os.Setenv("JWT_ALGORITHM", jwtAlgorithm)
	os.Setenv("JWT_USER_KEYID", jwtUserKeyID)
	os.Setenv("JWT_USER_SECRETKEY", jwtUserSecretkey)
	os.Setenv("JWT_SERVICE_KEYID", jwtServiceKeyID)
	os.Setenv("JWT_SERVICE_SECRETKEY", jwtServiceSecretkey)
	os.Setenv("REDIS_MAX_IDLE", redisMaxIdle)
	os.Setenv("REDIS_MAX_ACTIVE", redisMaxActive)
	os.Setenv("REDIS_PORT", redisPort)
	os.Setenv("REDIS_ADDRESS", redisAddress)

	conf, _ := SetConfigs()

	svcs := conf.SVCGetServices()

	tests.MaybeFail("Test_default_configs",
		tests.Expect(conf.GlbGetenv(), "development"),
		tests.Expect(conf.GlbGetServiceName(), serviceName),
		tests.Expect(svcs["frontend"].SVCGetID(), "222222"),
		tests.Expect(svcs["frontend"].SVCGetSecret(), "22222222"),
		tests.Expect(svcs["frontend"].SVCGetDomain(), "http://localhost:80"),
		tests.Expect(svcs["broker_svc"].SVCGetID(), "brokerSvc"),
		tests.Expect(svcs["registry_svc"].SVCGetDomain(), "http://localhost:4000"),
		tests.Expect(conf.MgoGetAuthDatabaseName(), mongoAuthDatabaseName),
		tests.Expect(conf.MgoGetUsersDatabaseName(), mongoUsersDatabaseName),
		tests.Expect(conf.MgoGetReplicaSetName(), mongoReplicaSetName),
		tests.Expect(conf.MgoGetCluster(), true),
		tests.Expect(conf.MgoGetURL(), mongoURLDevelopment),
		tests.Expect(conf.MgoGetUsername(), mongoUsername),
		tests.Expect(conf.MgoGetPassword(), mongoPassword),
		tests.Expect(conf.RdsGetPort(), redisPort),
		tests.Expect(conf.RdsGetAddress(), redisAddress),
		tests.Expect(conf.RdsGetMaxIdle(), 809),
		tests.Expect(conf.RdsGetMaxActive(), 120009),
		tests.Expect(conf.JwtGetAlgorithm(), jwtAlgorithm),
		tests.Expect(conf.JwtGetUserKeyID(), jwtUserKeyID),
		tests.Expect(conf.JwtGetUserSecretkey(), jwtUserSecretkey),
		tests.Expect(conf.JwtGetServiceKeyID(), jwtServiceKeyID),
		tests.Expect(conf.JwtGetServiceSecretkey(), jwtServiceSecretkey),
	)
}
