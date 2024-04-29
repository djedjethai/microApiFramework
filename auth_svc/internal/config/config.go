package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	serviceNameDefault                = "auth"
	configFileNameDefault             = "services"
	configFileExtenssionDefault       = "yaml"
	configFilePathDefault             = "../configs/v1"
	mongoAuthDatabaseNameDefault      = "oauth2"
	mongoUsersDatabaseNameDefault     = "authentication"
	mongoReplicaSetNameDefault        = "myReplicaSet"
	mongoUsernameDefault              = "admin"
	mongoPasswordDefault              = "password"
	mongoURLDefault                   = "mongodb://127.0.0.1:27017"
	mongoURLDevelopment               = "mongodb://mongo:27017"
	mongoURLStaging                   = ""
	mongoURLProduction                = ""
	isMongoClusterDefault             = false
	jwtAlgorithmDefault               = "HS256"
	jwtUserKeyIDDefault               = "userKeyIDDefault"
	jwtUserSecretkeyDefault           = "userSecretkeyDefault"
	jwtServiceKeyIDDefault            = "serviceKeyIDDefault"
	jwtServiceSecretkeyDefault        = "serviceSecretKeyDefault"
	redisPortDefault                  = "6379"
	redisAddressDefault               = "redis"
	redisMaxIdleDefault           int = 80
	redisMaxActiveDefault         int = 12000
)

func SetConfigs() (*Config, error) {
	// Global
	golangEnv := os.Getenv("GOENV")
	srvName := os.Getenv("SERVICE_NAME")
	c := NewConfig(golangEnv, srvName)

	// set services from yaml config files
	configName := os.Getenv("CONFIG_FILE_NAME")
	configExt := os.Getenv("CONFIG_FILE_EXT")
	configPath := os.Getenv("CONFIG_FILE_PATH")
	svcs := NewServices(configName, configExt, configPath)
	err := svcs.loadServices()
	if err != nil {
		return c, err
	}
	c.services = svcs

	// MongoDB
	mongoUsername := os.Getenv("MONGO_USERNAME")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	isMongoCluster := os.Getenv("MONGO_CLUSTER")
	authDatabase := os.Getenv("MONGO_AUTH_DATABASE")
	usersDatabase := os.Getenv("MONGO_USERS_DATABASE")
	mongoReplicaSetName := os.Getenv("MONGO_REPLICASET_NAME")
	c.mgoSetCluster(isMongoCluster)
	c.mgoSetUsername(mongoUsername)
	c.mgoSetPassword(mongoPassword)
	c.mgoSetAuthDatabaseName(authDatabase)
	c.mgoSetUsersDatabaseName(usersDatabase)
	c.mgoSetReplicaSetName(mongoReplicaSetName)

	// Encryption
	jwtAlgorithm := os.Getenv("JWT_ALGORITHM")
	jwtUserKeyID := os.Getenv("JWT_USER_KEYID")
	jwtUserSecretkey := os.Getenv("JWT_USER_SECRETKEY")
	jwtServiceKeyID := os.Getenv("JWT_SERVICE_KEYID")
	jwtServiceSecretkey := os.Getenv("JWT_SERVICE_SECRETKEY")
	c.jwtSetAlgorithm(jwtAlgorithm)
	c.jwtSetUserKeyID(jwtUserKeyID)
	c.jwtSetUserSecretkey(jwtUserSecretkey)
	c.jwtSetServiceKeyID(jwtServiceKeyID)
	c.jwtSetServiceSecretkey(jwtServiceSecretkey)

	// Redis
	redisMaxIdle := os.Getenv("REDIS_MAX_IDLE")
	redisMaxActive := os.Getenv("REDIS_MAX_ACTIVE")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddress := os.Getenv("REDIS_ADDRESS")
	c.rdsSetMaxIdle(redisMaxIdle)
	c.rdsSetMaxActive(redisMaxActive)
	c.rdsSetPort(redisPort)
	c.rdsSetAddress(redisAddress)

	return c, nil
}

// Config wrap all configs
type Config struct {
	*Global
	*services
	*MongoDB
	*JWTEncryption
	*Redis
}

func NewConfig(goEnv string, serviceName ...string) *Config {
	srvName := serviceNameDefault
	if len(serviceName) > 0 && serviceName[0] != "" {
		srvName = serviceName[0]
	}
	g := NewGlobal(goEnv, srvName)

	c := &Config{
		Global:        g,
		MongoDB:       NewMongoDB(g.GlbGetenv()),
		JWTEncryption: NewJWTEncryption(),
		Redis:         NewRedis(g.GlbGetenv()),
	}

	return c
}

// Global may concern any configs
type Global struct {
	golangEnv   string
	serviceName string
}

func NewGlobal(golangEnv, svcName string) *Global {
	return &Global{golangEnv, svcName}
}

func (g *Global) GlbGetenv() string {
	return g.golangEnv
}

func (g *Global) GlbSetServiceName(srvName string) {
	if srvName != "" {
		g.serviceName = srvName
	}
}

func (g *Global) GlbGetServiceName() string {
	return g.serviceName
}

// services
type IService interface {
	SVCGetID() string
	SVCGetSecret() string
	SVCGetDomain() string
}

type service struct {
	id     string
	secret string
	domain string
}

func (s service) SVCGetID() string {
	return s.id
}
func (s service) SVCGetSecret() string {
	return s.secret
}
func (s service) SVCGetDomain() string {
	return s.domain
}

type services struct {
	services             map[string]IService
	configFileName       string
	configFileExtenssion string
	configFilePath       string
}

func NewServices(configName, configExtenssion, path string) *services {
	svcs := &services{}
	svcs.services = make(map[string]IService)
	svcs.configFileName = configFileNameDefault
	svcs.configFileExtenssion = configFileExtenssionDefault
	svcs.configFilePath = configFilePathDefault

	if configName != "" {
		svcs.configFileName = configName
	}
	if configExtenssion != "" {
		svcs.configFileExtenssion = configExtenssion
	}
	if path != "" {
		svcs.configFilePath = path
	}

	return svcs
}

func (ss *services) loadServices() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(currentDir, ss.configFilePath)

	viper.SetConfigName(ss.configFileName) // Config file name without extension
	viper.SetConfigType(ss.configFileExtenssion)
	viper.AddConfigPath(configFilePath)

	// Read the configuration file
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Get all keys, then add all services infos to the configs
	keys := viper.AllKeys()
	ref := make(map[string]struct{})
	for i := 0; i < len(keys); i++ {
		parts := strings.Split(keys[i], ".")
		if _, ok := ref[parts[0]]; !ok {
			ref[parts[0]] = struct{}{}
			services := viper.Sub(parts[0])
			settings := services.AllSettings()
			svc := service{}
			for key, value := range settings {
				switch key {
				case "id":
					svc.id = value.(string)
				case "secret":
					svc.secret = value.(string)
				case "domain":
					svc.domain = value.(string)
				}
			}
			ss.services[parts[0]] = svc
		}
	}
	return nil
}

func (ss *services) SVCGetServices() map[string]IService {
	return ss.services
}

// MongoDB are the MongoDB configs
type MongoDB struct {
	golangEnv         string
	isMongoCluster    bool
	mongoURL          string
	mongoUsername     string
	mongoPassword     string
	authDatabaseName  string
	usersDatabaseName string
	replicaSetName    string
}

func NewMongoDB(goEnv string) *MongoDB {
	m := &MongoDB{golangEnv: goEnv}
	m.mgoSetURL()
	m.isMongoCluster = isMongoClusterDefault
	m.mongoUsername = mongoUsernameDefault
	m.mongoPassword = mongoPasswordDefault
	m.authDatabaseName = mongoAuthDatabaseNameDefault
	m.usersDatabaseName = mongoUsersDatabaseNameDefault
	m.replicaSetName = mongoReplicaSetNameDefault
	return m
}

func (m *MongoDB) mgoSetAuthDatabaseName(name string) {
	if name != "" {
		m.authDatabaseName = name
	}
}

func (m *MongoDB) MgoGetAuthDatabaseName() string {
	return m.authDatabaseName
}

func (m *MongoDB) mgoSetUsersDatabaseName(name string) {
	if name != "" {
		m.usersDatabaseName = name
	}
}

func (m *MongoDB) MgoGetUsersDatabaseName() string {
	return m.usersDatabaseName
}

func (m *MongoDB) mgoSetReplicaSetName(name string) {
	if name != "" {
		m.replicaSetName = name
	}
}

func (m *MongoDB) MgoGetReplicaSetName() string {
	return m.replicaSetName
}

func (m *MongoDB) mgoSetCluster(isMongoCluster string) {
	m.isMongoCluster = false
	if isMongoCluster != "" && isMongoCluster == "true" {
		m.isMongoCluster = true
	}
}

func (m *MongoDB) MgoGetCluster() bool {
	return m.isMongoCluster
}

func (m *MongoDB) mgoSetURL() {
	switch m.golangEnv {
	case "development":
		m.mongoURL = mongoURLDevelopment
	case "staging":
		m.mongoURL = mongoURLStaging
	case "production":
		m.mongoURL = mongoURLProduction
	default:
		m.mongoURL = mongoURLDefault
	}
}

func (m *MongoDB) MgoGetURL() string {
	return m.mongoURL
}

func (m *MongoDB) mgoSetUsername(mongoUsername string) {
	if mongoUsername != "" {
		m.mongoUsername = mongoUsername
	}
}

func (m *MongoDB) MgoGetUsername() string {
	return m.mongoUsername
}

func (m *MongoDB) mgoSetPassword(mongoPassword string) {
	if mongoPassword != "" {
		m.mongoPassword = mongoPassword
	}
}

func (m *MongoDB) MgoGetPassword() string {
	return m.mongoPassword
}

// Redis
type Redis struct {
	golangEnv    string
	redisPort    string
	redisAddress string
	maxIdle      int
	maxActive    int
}

func NewRedis(goEnv string) *Redis {
	r := &Redis{golangEnv: goEnv}
	r.redisPort = redisPortDefault
	r.redisAddress = redisAddressDefault
	r.maxIdle = redisMaxIdleDefault
	r.maxActive = redisMaxActiveDefault

	return r
}

func (r *Redis) rdsSetPort(p string) {
	if p != "" {
		r.redisPort = p
	}
}

func (r *Redis) RdsGetPort() string {
	return r.redisPort
}

func (r *Redis) rdsSetAddress(a string) {
	if a != "" {
		r.redisAddress = a
	}
}

func (r *Redis) RdsGetAddress() string {
	return r.redisAddress
}

func (r *Redis) rdsSetMaxIdle(mi string) {
	num, err := strconv.Atoi(mi)
	if err == nil && num > 0 {
		r.maxIdle = num
	}
}

func (r *Redis) RdsGetMaxIdle() int {
	return r.maxIdle
}

func (r *Redis) rdsSetMaxActive(ma string) {
	num, err := strconv.Atoi(ma)
	if err == nil && num > 0 {
		r.maxActive = num
	}
}

func (r *Redis) RdsGetMaxActive() int {
	return r.maxActive
}

// Encryption are Encryption configs
type JWTEncryption struct {
	algorithm        string
	userKeyID        string
	userSecretkey    string
	serviceKeyID     string
	serviceSecretkey string
}

func NewJWTEncryption() *JWTEncryption {
	e := &JWTEncryption{}
	e.algorithm = jwtAlgorithmDefault
	e.userKeyID = jwtUserKeyIDDefault
	e.userSecretkey = jwtUserSecretkeyDefault
	e.serviceKeyID = jwtServiceKeyIDDefault
	e.serviceSecretkey = jwtServiceSecretkeyDefault
	return e
}

func (e *JWTEncryption) jwtSetAlgorithm(algo string) {
	if algo != "" {
		e.algorithm = algo
	}
}

func (e *JWTEncryption) JwtGetAlgorithm() string {
	return e.algorithm
}

func (e *JWTEncryption) jwtSetUserKeyID(ukid string) {
	if ukid != "" {
		e.userKeyID = ukid
	}
}

func (e *JWTEncryption) JwtGetUserKeyID() string {
	return e.userKeyID
}

func (e *JWTEncryption) jwtSetUserSecretkey(usk string) {
	if usk != "" {
		e.userSecretkey = usk
	}
}

func (e *JWTEncryption) JwtGetUserSecretkey() string {
	return e.userSecretkey
}

func (e *JWTEncryption) jwtSetServiceKeyID(skid string) {
	if skid != "" {
		e.serviceKeyID = skid
	}
}

func (e *JWTEncryption) JwtGetServiceKeyID() string {
	return e.serviceKeyID
}

func (e *JWTEncryption) jwtSetServiceSecretkey(ssk string) {
	if ssk != "" {
		e.serviceSecretkey = ssk
	}
}

func (e *JWTEncryption) JwtGetServiceSecretkey() string {
	return e.serviceSecretkey
}
