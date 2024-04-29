package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// SetVarEnv Set the default varEnv and overwrite it if some exist
func SetVarEnv() {
	// global
	viper.SetDefault("GOENV", "localhost")
	viper.SetDefault("SERVICE_NAME", "registrySvc")

	// http
	viper.SetDefault("HTTP_ADDRESS", "localhost")
	viper.SetDefault("HTTP_PORT", "4000")
	viper.SetDefault("PATH_STORAGE", "..")
	viper.SetDefault("GRPC_DIRECTORY", "api")
	viper.SetDefault("CONFIGS_DIRECTORY", "configs")

	// JWTRequestConfig
	viper.SetDefault("AUTH_SVC_URL", "http://localhost:9096/v1")
	viper.SetDefault("AUTH_SVC_PATH", "apiauth")
	viper.SetDefault("AUTH_SVC_TOKEN_ENDPOINT", "http://localhost:9096/v1/oauth/token")
	viper.SetDefault("CODE_VERIFIER", "exampleCodeVerifier")
	viper.SetDefault("SERVICE_KEY_ID", "registrySvc")
	viper.SetDefault("SERVICE_SECRET_KEY", "registrySvcSecret")
	viper.SetDefault("SCOPE", "read, openid")

	// Read configuration from environment variables or configuration files if needed
	viper.AutomaticEnv()
}

// set grpcPorts
type Config struct {
	*Global
	*Http
	*JWTRequestConfig
}

// func NewConfigs(goEnv string, serviceName string) *Configs {
func NewConfig() *Config {
	// fmt.Println("NewConfig - see env: ", goEnv)
	g := NewGlobal()

	http := NewHttp()

	jwtRequestConfig := NewJWTRequestConfig()

	c := &Config{
		Global:           g,
		Http:             http,
		JWTRequestConfig: jwtRequestConfig,
	}

	return c
}

// JWTRequestConfig are the configuration to get the auth's JWToken
type JWTRequestConfig struct {
	authSvcURL           string
	authSvcPath          string
	authSvcTokenEndpoint string
	codeVerifier         string
	serviceKeyID         string
	serviceSecretkey     string
	scope                string
}

func NewJWTRequestConfig() *JWTRequestConfig {
	jrc := &JWTRequestConfig{}
	jrc.authSvcURL = viper.GetString("AUTH_SVC_URL")
	jrc.authSvcPath = viper.GetString("AUTH_SVC_PATH")
	jrc.authSvcTokenEndpoint = viper.GetString("AUTH_SVC_TOKEN_ENDPOINT")
	jrc.codeVerifier = viper.GetString("CODE_VERIFIER")
	jrc.serviceKeyID = viper.GetString("SERVICE_KEY_ID")
	jrc.serviceSecretkey = viper.GetString("SERVICE_SECRET_KEY")
	jrc.scope = viper.GetString("SCOPE")

	return jrc
}

func (r *JWTRequestConfig) JwtGetAuthSvcURL() string {
	return r.authSvcURL
}

func (r *JWTRequestConfig) JwtGetAuthSvcPath() string {
	return r.authSvcPath
}

func (r *JWTRequestConfig) JwtGetAuthSvcTokenEndpoint() string {
	return r.authSvcTokenEndpoint
}

func (r *JWTRequestConfig) JwtGetCodeVerifier() string {
	return r.codeVerifier
}

func (r *JWTRequestConfig) JwtGetServiceKeyID() string {
	return r.serviceKeyID
}

func (r *JWTRequestConfig) JwtGetServiceSecretKey() string {
	return r.serviceSecretkey
}

func (r *JWTRequestConfig) JwtGetScope() string {
	return r.scope
}

// Global may concern any configs
type Global struct {
	golangEnv   string
	serviceName string
}

func NewGlobal() *Global {
	g := &Global{}
	g.golangEnv = viper.GetString("GOENV")
	g.serviceName = viper.GetString("SERVICE_NAME")

	return g
}

func (g *Global) GLBGetenv() string {
	return g.golangEnv
}

func (g *Global) GLBGetServiceName() string {
	return g.serviceName
}

// HTTTP configs
type Http struct {
	address string
	port    string
}

func NewHttp() *Http {
	h := &Http{}
	h.address = viper.GetString("HTTP_ADDRESS")
	h.port = viper.GetString("HTTP_PORT")

	return h
}

func (h *Http) HTTPGetAddress() string {
	return h.address
}

func (h *Http) HTTPGetPort() string {
	return h.port
}

func (h *Http) HTTPGetHTTPFormatedURL() string {
	return fmt.Sprintf("http://%v:%v", h.address, h.port)
}

/*********
* Rest are the configurations for the rest API
*********/
// IRestConfigs is the interface to the Rest configurations
type IRestConfig interface {
	RESTSetPathToStorage(pts string)
	RESTGetPathToStorage() string
	RESTSetGrpcDirectory(grpcDir string)
	RESTGetGrpcDirectory() string
	RESTSetConfigsDirectory(confDir string)
	RESTGetConfigsDirectory() string
}

// Rest hold the Rest configurations
type RestConfig struct {
	pathToStorage    string
	grpcDirectory    string
	configsDirectory string
}

func NewRestConfig() *RestConfig {
	rc := &RestConfig{}
	rc.pathToStorage = viper.GetString("PATH_STORAGE")
	rc.grpcDirectory = viper.GetString("GRPC_DIRECTORY")
	rc.configsDirectory = viper.GetString("CONFIGS_DIRECTORY")

	return rc
}

func (r *RestConfig) RESTSetPathToStorage(pts string) {
	if len(pts) > 0 {
		r.pathToStorage = pts
	}
}

func (r *RestConfig) RESTGetPathToStorage() string {
	return r.pathToStorage
}

func (r *RestConfig) RESTSetGrpcDirectory(grpcDir string) {
	if len(grpcDir) > 0 {
		r.grpcDirectory = grpcDir
	}
}

func (r *RestConfig) RESTGetGrpcDirectory() string {
	return r.grpcDirectory
}

func (r *RestConfig) RESTSetConfigsDirectory(confDir string) {
	if len(confDir) > 0 {
		r.configsDirectory = confDir
	}
}
func (r *RestConfig) RESTGetConfigsDirectory() string {
	return r.configsDirectory
}
