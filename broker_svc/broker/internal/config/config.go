package config

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strconv"
)

const (
	// server
	GOENVDefault   string = "localhost"
	svcNameDefault string = "brokerSvc"

	// services that http req will be proxy to
	configFileNameDefault       = "config"
	configFileExtenssionDefault = "yaml"
	configFilePathDefault       = "../broker/configs/v1/servicesAddr/"

	// http
	httpAddressDefault string = "localhost"
	httpPortDefault           = "8080"

	// grpc
	grpcAddressDefault       = "localhost" // !!! 127.0.0.1 does not work
	jwtValidationPortDefault = "50002"

	// tls
	pathToTLSDefault      string = "../broker/configs/v1/certificates"
	clientCertFileDefault string = "client.crt"
	clientKeyFileDefault  string = "client.key"
	serverCertFileDefault string = "server.crt"
	serverKeyFileDefault  string = "server.key"
	caFileDefault         string = "rootCA.crt"

	// JWTrequestConfigs
	authSvcURLDefault           string = "http://localhost:9096/v1" // all varEnv
	authSvcPathDefault                 = "apiauth"
	authSvcTokenEndpointDefault        = "http://localhost:9096/v1/oauth/token"
	codeVerifierDefault                = "exampleCodeVerifier"
	serviceKeyIDDefault                = "brokerSvc"
	serviceSecretkeyDefault            = "brokerSvcSecret"
	scopeDefault                       = "read, openid"

	// TOKENservice
	signedKeyDefault string = "mySecretKey"

	// observability
	obsSamplingDefault          float64 = 0.6
	obsScratchDelayDefault      int     = 30
	obsCollectorEndpointDefault string  = "otel_collector:4317"
)

func SetConfigs() (*Config, error) {

	golangEnv := os.Getenv("GOENV")
	srvName := os.Getenv("SERVICE_NAME")
	c := NewConfig(golangEnv, srvName)

	// set services from yaml config files
	configName := os.Getenv("CONFIG_FILE_NAME")
	configExt := os.Getenv("CONFIG_FILE_EXT")
	configPath := os.Getenv("CONFIG_FILE_PATH")
	svcs := NewServices(configName, configExt, configPath)
	err := svcs.loadServices(c.GlbGetenv())
	if err != nil {
		return c, err
	}
	c.services = svcs

	// set http configs
	svcAddr := os.Getenv("SERVICE_ADDRESS")
	svcPort := os.Getenv("SERVICE_PORT")
	c.httpSetAddress(svcAddr)
	c.httpSetPort(svcPort)

	// set grpc configs
	grpcAddr := os.Getenv("GRPC_ADDRESS")
	grpcJwtValidationPort := os.Getenv("GRPC_JWT_VALIDATION_PORT")
	c.grpcSetAddress(grpcAddr)
	c.grpcSetJwtValidationPort(grpcJwtValidationPort)

	// set the JWTRequestConfig
	authSvcURL := os.Getenv("AUTH_SVC_URL")
	authSvcPATH := os.Getenv("AUTH_SVC_PATH")
	authSvcTokenEndpoint := os.Getenv("AUTH_SVC_TOKEN_ENDPOINT")
	codeVerifier := os.Getenv("CODE_VERIFIER")
	serviceKeyID := os.Getenv("SERVICE_KEY_ID")
	serviceSecretkey := os.Getenv("SERVICE_SECRET_KEY")
	scope := os.Getenv("SCOPE")
	c.jwtSetAuthSvcURL(authSvcURL)
	c.jwtSetAuthSvcPath(authSvcPATH)
	c.jwtSetAuthSvcTokenEndpoint(authSvcTokenEndpoint)
	c.jwtSetCodeVerifier(codeVerifier)
	c.jwtSetServiceKeyID(serviceKeyID)
	c.jwtSetServiceSecretKey(serviceSecretkey)
	c.jwtSetScope(scope)

	// set configs related to jwtTokenValidation
	// TODO in prod must get it from secrets
	signedKey := os.Getenv("SIGNED_KEY")
	c.tokenSetSignedKey(signedKey)

	// set the observability
	sampling := os.Getenv("OBS_SAMPLING")
	scrDelay := os.Getenv("OBS_SCRATCH_DELAY")
	collEndpoint := os.Getenv("OBS_COLLECTOR_ENDPOINT")
	c.obsSetSampling(sampling)
	c.obsSetScratchDelay(scrDelay)
	c.obsSetCollectorEndpoint(collEndpoint)

	// set the TLS configs
	clientCertFile := os.Getenv("CLIENT_CERT_FILE")
	clientKeyFile := os.Getenv("CLIENT_KEY_FILE")
	serverCertFile := os.Getenv("SERVER_CERT_FILE")
	serverKeyFile := os.Getenv("SERVER_KEY_FILE")
	caFile := os.Getenv("CA_FILE")
	pathToTLS := os.Getenv("PATH_TO_TLS")
	err = setTLSConfig(
		c,
		clientCertFile,
		clientKeyFile,
		serverCertFile,
		serverKeyFile,
		caFile,
		pathToTLS,
	)

	return c, err
}

// Services
type Service interface {
	SVCGetAddress() string
	SVCGetPort() string
}

type service struct {
	address string
	port    string
}

func (s service) SVCGetAddress() string {
	return s.address
}

func (s service) SVCGetPort() string {
	return s.port
}

type services struct {
	services             map[string]service
	configFileName       string
	configFileExtenssion string
	configFilePath       string
}

func NewServices(configName, configExtenssion, path string) *services {
	svcs := &services{}
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

func (ss *services) loadServices(env string) error {
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

	var sv service
	var ok bool
	var addressValue string
	var addressPort string
	var valueMap map[string]interface{}

	services := viper.Sub(env)
	settings := services.AllSettings()
	var svs = make(map[string]service, len(settings))
	if services != nil {
		for key, value := range settings {
			// Type assertion for the map
			valueMap, ok = value.(map[string]interface{})
			if !ok {
				fmt.Println("Error: Unable to assert value as map[string]interface{}")
				return fmt.Errorf("err assert value...")
			}

			addressValue, ok = valueMap["address"].(string)
			if !ok {
				return fmt.Errorf("Error: Unable to retrieve address value")
			}

			addressPort, ok = valueMap["port"].(string)
			if !ok {
				return fmt.Errorf("Error: Unable to retrieve port value")
			}

			sv = service{
				address: addressValue,
				port:    addressPort,
			}

			svs[key] = sv
		}
		ss.services = svs
	}

	return nil
}

func (ss *services) SVCSGetServices() map[string]service {
	return ss.services
}

// Config wrap all configs
type Config struct {
	*global
	*http
	*grpc
	*jwtRequestConfig
	*token
	ClientTLSConfig *tls.Config
	ServerTLSConfig *tls.Config
	*services
	*observability
}

func NewConfig(goEnv string, serviceName ...string) *Config {
	srvName := svcNameDefault
	if len(serviceName[0]) > 0 {
		srvName = serviceName[0]
	}

	env := GOENVDefault
	if len(goEnv) > 0 {
		env = goEnv
	}

	g := NewGlobal(env, srvName)

	grpc := NewGrpc()
	http := NewHttp()

	c := &Config{
		global:           g,
		http:             http,
		grpc:             grpc,
		jwtRequestConfig: NewJWTRequestConfig(),
		token:            NewToken(),
		observability:    NewObservability(),
	}

	return c
}

// Global may concern any configs
type global struct {
	golangEnv   string
	serviceName string
	jwtoken     string
}

func NewGlobal(golangEnv, svcName string) *global {
	return &global{
		golangEnv:   golangEnv,
		serviceName: svcName}
}

func (g *global) GlbGetenv() string {
	return g.golangEnv
}

func (g *global) GlbSetJWToken(tkn string) {
	g.jwtoken = tkn
}

func (g *global) GlbGetJWToken() string {
	return g.jwtoken
}

func (g *global) GLBSetServiceName(srvName string) {
	if srvName != "" {
		g.serviceName = srvName
	}
}

func (g *global) GlbGetServiceName() string {
	return g.serviceName
}

// HTTTP configs
type http struct {
	address string
	port    string
}

func NewHttp() *http {
	h := &http{}

	// set default
	h.address = httpAddressDefault
	h.port = httpPortDefault

	return h
}

func (h *http) httpSetAddress(addr string) {
	if addr != "" {
		h.address = addr
	}
}

func (h *http) HTTPGetAddress() string {
	return h.address
}

func (h *http) httpSetPort(port string) {
	if port != "" {
		h.port = port
	}
}

func (h *http) HTTPGetPort() string {
	return h.port
}

func (h *http) HTTPGetHTTPFormatedURL() string {
	return fmt.Sprintf("http://%v:%v", h.address, h.port)
}

// GRPC configs
type grpc struct {
	grpcAddress       string
	jwtValidationPort string
}

func NewGrpc() *grpc {
	g := &grpc{}

	// set default
	g.grpcAddress = grpcAddressDefault
	g.jwtValidationPort = jwtValidationPortDefault

	return g
}

func (g *grpc) grpcSetAddress(address string) {
	if address != "" {
		g.grpcAddress = address
	}
}
func (g *grpc) GRPCGetAddress() string {
	return g.grpcAddress
}

func (g *grpc) grpcSetJwtValidationPort(port string) {
	if port != "" {
		g.jwtValidationPort = port
	}
}
func (g *grpc) GRPCGetJwtValidationPort() string {
	return g.jwtValidationPort
}

// setTLSConfig set the tls configuration
func setTLSConfig(c *Config, ccf, ckf, scf, skf, ca, pathToTLS string) error {
	// client tls config
	var clientCertFile = clientCertFileDefault
	var clientKeyFile = clientKeyFileDefault
	var caFile = caFileDefault
	if ccf != "" {
		clientCertFile = ccf
	}
	if ckf != "" {
		clientKeyFile = ckf
	}
	if ca != "" {
		caFile = ca
	}
	cltTLSConfig := tlsConfig{
		certFile:      configFileTLS(pathToTLS, clientCertFile),
		keyFile:       configFileTLS(pathToTLS, clientKeyFile),
		caFile:        configFileTLS(pathToTLS, caFile),
		serverAddress: c.GRPCGetAddress(),
	}
	cltTLSConfig.server = false

	// server tls config
	var serverCertFile = serverCertFileDefault
	var serverKeyFile = serverKeyFileDefault
	if scf != "" {
		serverCertFile = scf
	}
	if skf != "" {
		serverKeyFile = skf
	}
	srvTLSConfig := tlsConfig{
		certFile:      configFileTLS(pathToTLS, serverCertFile),
		keyFile:       configFileTLS(pathToTLS, serverKeyFile),
		caFile:        configFileTLS(pathToTLS, caFile),
		serverAddress: c.GRPCGetAddress(),
	}
	srvTLSConfig.server = true

	var err error
	c.ClientTLSConfig, err = setupTLSConfig(cltTLSConfig)
	if err != nil {
		fmt.Println("error setting ClientTLSConfig")
		return err
	}
	c.ServerTLSConfig, err = setupTLSConfig(srvTLSConfig)
	if err != nil {
		fmt.Println("error setting ServerTLSConfig")
		return err
	}
	return nil
}

// JWTRequestConfig are the configuration to get the auth's JWToken
type jwtRequestConfig struct {
	authSvcURL           string
	authSvcPath          string
	authSvcTokenEndpoint string
	codeVerifier         string
	serviceKeyID         string
	serviceSecretkey     string
	scope                string
}

func NewJWTRequestConfig() *jwtRequestConfig {
	r := &jwtRequestConfig{}
	r.authSvcURL = authSvcURLDefault
	r.authSvcPath = authSvcPathDefault
	r.authSvcTokenEndpoint = authSvcTokenEndpointDefault
	r.codeVerifier = codeVerifierDefault
	r.serviceKeyID = serviceKeyIDDefault
	r.serviceSecretkey = serviceSecretkeyDefault
	r.scope = scopeDefault
	return r
}

func (r *jwtRequestConfig) jwtSetAuthSvcURL(v string) {
	if v != "" {
		r.authSvcURL = v
	}
}
func (r *jwtRequestConfig) JwtGetAuthSvcURL() string {
	return r.authSvcURL
}

func (r *jwtRequestConfig) jwtSetAuthSvcPath(v string) {
	if v != "" {
		r.authSvcPath = v
	}
}
func (r *jwtRequestConfig) JwtGetAuthSvcPath() string {
	return r.authSvcPath
}

func (r *jwtRequestConfig) jwtSetAuthSvcTokenEndpoint(v string) {
	if v != "" {
		r.authSvcTokenEndpoint = v
	}
}
func (r *jwtRequestConfig) JwtGetAuthSvcTokenEndpoint() string {
	return r.authSvcTokenEndpoint
}

func (r *jwtRequestConfig) jwtSetCodeVerifier(v string) {
	if v != "" {
		r.codeVerifier = v
	}
}
func (r *jwtRequestConfig) JwtGetCodeVerifier() string {
	return r.codeVerifier
}

func (r *jwtRequestConfig) jwtSetServiceKeyID(v string) {
	if v != "" {
		r.serviceKeyID = v
	}
}
func (r *jwtRequestConfig) JwtGetServiceKeyID() string {
	return r.serviceKeyID
}

func (r *jwtRequestConfig) jwtSetServiceSecretKey(v string) {
	if v != "" {
		r.serviceSecretkey = v
	}
}
func (r *jwtRequestConfig) JwtGetServiceSecretKey() string {
	return r.serviceSecretkey
}

func (r *jwtRequestConfig) jwtSetScope(v string) {
	if v != "" {
		r.scope = v
	}
}
func (r *jwtRequestConfig) JwtGetScope() string {
	return r.scope
}

// token handle jwtToken validation
type token struct {
	signedKey string
}

func NewToken() *token {
	return &token{signedKeyDefault}
}

func (t *token) tokenSetSignedKey(signedKey string) {
	if signedKey != "" {
		t.signedKey = signedKey
	}
}

func (t *token) TOKENGetSignedKey() string {
	return t.signedKey
}

// observability handle the observability configs
type observability struct {
	obsSampling          float64
	obsScratchDelay      int
	obsCollectorEndpoint string
}

func NewObservability() *observability {
	obs := &observability{}

	obs.obsSampling = obsSamplingDefault
	obs.obsScratchDelay = obsScratchDelayDefault
	obs.obsCollectorEndpoint = obsCollectorEndpointDefault

	return obs
}

func (o *observability) obsSetSampling(sampling string) error {
	if sampling != "" {
		// Convert string to float64
		f64, err := strconv.ParseFloat(sampling, 64)
		if err != nil {
			return fmt.Errorf("error converting sampling, str to float64: %v", err)
		}
		o.obsSampling = f64
	}
	return nil
}

func (o *observability) OBSGetSampling() float64 {
	return o.obsSampling
}

func (o *observability) obsSetScratchDelay(delay string) error {
	if delay != "" {
		// Convert string to int
		i, err := strconv.Atoi(delay)
		if err != nil {
			return fmt.Errorf("error converting scratchDelay, str to int: %v", err)
		}
		o.obsScratchDelay = i
	}
	return nil
}

func (o *observability) OBSGetScratchDelay() int {
	return o.obsScratchDelay
}

func (o *observability) obsSetCollectorEndpoint(ep string) error {
	if ep != "" {
		o.obsCollectorEndpoint = ep
	}
	return nil
}

func (o *observability) OBSGetCollectorEndpoint() string {
	return o.obsCollectorEndpoint
}
