package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
)

const (
	// server
	GOENVDefault   string = "localhost"
	svcNameDefault string = "order"

	jwtValidationPortDefault    string = "50002"
	jwtValidationAddressDefault string = "127.0.0.1"

	// grpc
	grpcAddressDefault string = "localhost"
	grpcPortDefault           = "50001"

	// tls
	pathToTLSDefault      string = "../order/configs/v1/certificates"
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
	serviceKeyIDDefault                = "order"
	serviceSecretkeyDefault            = "orderSecret"
	scopeDefault                       = "read, openid"

	// observability
	obsSamplingDefault          float64 = 0.6
	obsScratchDelayDefault      int     = 30
	obsCollectorEndpointDefault string  = "otel_collector:4317"

	// kafka
	kafkaURLDefault                  string = "127.0.0.1:29093"
	kafkaProducerKeyLocationDefault         = "../order/configs/v1/kafka/producer.key.pem"
	kafkaProducerCertLocationDefault        = "../order/configs/v1/kafka/producer.cer.pem"
	kafkaConsumerKeyLocationDefault         = "../order/configs/v1/kafka/consumer.key.pem"
	kafkaConsumerCertLocationDefault        = "../order/configs/v1/kafka/consumer.cer.pem"
	kafkaPassphraseDefault                  = "datahub"
	kafkaBrokerCertLocationDefault          = "../order/configs/v1/kafka/broker.cer.pem"
)

func SetConfigs() (*Config, error) {

	golangEnv := os.Getenv("GOENV")
	srvName := os.Getenv("SERVICE_NAME")
	c := NewConfig(golangEnv, srvName)

	// set jwtValidation
	jwtvAddr := os.Getenv("JWT_VALIDATION_ADDRESS")
	jwtvPort := os.Getenv("JWT_VALIDATION_PORT")
	c.JWTVSetAddress(jwtvAddr)
	c.JWTVSetPort(jwtvPort)

	// set grpc configs
	svcAddr := os.Getenv("SERVICE_ADDRESS")
	svcPort := os.Getenv("SERVICE_PORT")
	c.GRPCSetAddress(svcAddr)
	c.GRPCSetPort(svcPort)

	// set the TLS configs
	clientCertFile := os.Getenv("CLIENT_CERT_FILE")
	clientKeyFile := os.Getenv("CLIENT_KEY_FILE")
	serverCertFile := os.Getenv("SERVER_CERT_FILE")
	serverKeyFile := os.Getenv("SERVER_KEY_FILE")
	caFile := os.Getenv("CA_FILE")
	pathToTLS := os.Getenv("PATH_TO_TLS")
	err := setTLSConfig(
		c,
		clientCertFile,
		clientKeyFile,
		serverCertFile,
		serverKeyFile,
		caFile,
		pathToTLS,
	)
	if err != nil {
		return c, err
	}

	// set the JWTRequestConfig
	authSvcURL := os.Getenv("AUTH_SVC_URL")
	authSvcPATH := os.Getenv("AUTH_SVC_PATH")
	authSvcTokenEndpoint := os.Getenv("AUTH_SVC_TOKEN_ENDPOINT")
	codeVerifier := os.Getenv("CODE_VERIFIER")
	serviceKeyID := os.Getenv("SERVICE_KEY_ID")
	serviceSecretkey := os.Getenv("SERVICE_SECRET_KEY")
	scope := os.Getenv("SCOPE")
	c.JwtSetAuthSvcURL(authSvcURL)
	c.JwtSetAuthSvcPath(authSvcPATH)
	c.JwtSetAuthSvcTokenEndpoint(authSvcTokenEndpoint)
	c.JwtSetCodeVerifier(codeVerifier)
	c.JwtSetServiceKeyID(serviceKeyID)
	c.JwtSetServiceSecretKey(serviceSecretkey)
	c.JwtSetScope(scope)

	// set the observability
	sampling := os.Getenv("OBS_SAMPLING")
	scrDelay := os.Getenv("OBS_SCRATCH_DELAY")
	collEndpoint := os.Getenv("OBS_COLLECTOR_ENDPOINT")
	c.obsSetSampling(sampling)
	c.obsSetScratchDelay(scrDelay)
	c.obsSetCollectorEndpoint(collEndpoint)

	// kafka
	kafkaURL := os.Getenv("KAFKA_URL")
	kafkaProdKeyLocation := os.Getenv("KAFKA_PRODUCER_KEY_LOCATION")
	kafkaProdCertLocation := os.Getenv("KAFKA_PRODUCER_CERT_LOCATION")
	kafkaConsKeyLocation := os.Getenv("KAFKA_CONSUMER_KEY_LOCATION")
	kafkaConsCertLocation := os.Getenv("KAFKA_CONSUMER_CERT_LOCATION")
	kafkaPassphrase := os.Getenv("KAFKA_PASSPHRASE")
	kafkaBrokerCertLocation := os.Getenv("KAFKA_BROKER_CERT_LOCATION")
	c.kfkSetURL(kafkaURL)
	c.kfkSetProducerKeyLocation(kafkaProdKeyLocation)
	c.kfkSetProducerCertLocation(kafkaProdCertLocation)
	c.kfkSetConsumerKeyLocation(kafkaConsKeyLocation)
	c.kfkSetConsumerCertLocation(kafkaConsCertLocation)
	c.kfkSetPassphrase(kafkaPassphrase)
	c.kfkSetBrokerCertLocation(kafkaBrokerCertLocation)

	return c, nil
}

type kafka struct {
	kafkaURL                  string
	kafkaProducerKeyLocation  string
	kafkaProducerCertLocation string
	kafkaConsumerKeyLocation  string
	kafkaConsumerCertLocation string
	kafkaPassphrase           string
	kafkaBrokerCertLocation   string
}

func NewKafka() *kafka {
	return &kafka{
		kafkaURL:                  kafkaURLDefault,
		kafkaProducerKeyLocation:  kafkaProducerKeyLocationDefault,
		kafkaProducerCertLocation: kafkaProducerCertLocationDefault,
		kafkaConsumerKeyLocation:  kafkaConsumerKeyLocationDefault,
		kafkaConsumerCertLocation: kafkaConsumerCertLocationDefault,
		kafkaPassphrase:           kafkaPassphraseDefault,
		kafkaBrokerCertLocation:   kafkaBrokerCertLocationDefault,
	}
}

func (k *kafka) kfkSetURL(url string) {
	if url != "" {
		k.kafkaURL = url
	}
}
func (k *kafka) KFKGetURL() string {
	return k.kafkaURL
}

func (k *kafka) kfkSetProducerKeyLocation(kl string) {
	if kl != "" {
		k.kafkaProducerKeyLocation = kl
	}
}
func (k *kafka) KFKGetProducerKeyLocation() string {
	return k.kafkaProducerKeyLocation
}

func (k *kafka) kfkSetProducerCertLocation(cl string) {
	if cl != "" {
		k.kafkaProducerCertLocation = cl
	}
}
func (k *kafka) KFKGetProducerCertLocation() string {
	return k.kafkaProducerCertLocation
}

func (k *kafka) kfkSetConsumerKeyLocation(kl string) {
	if kl != "" {
		k.kafkaConsumerKeyLocation = kl
	}
}
func (k *kafka) KFKGetConsumerKeyLocation() string {
	return k.kafkaConsumerKeyLocation
}

func (k *kafka) kfkSetConsumerCertLocation(cl string) {
	if cl != "" {
		k.kafkaConsumerCertLocation = cl
	}
}
func (k *kafka) KFKGetConsumerCertLocation() string {
	return k.kafkaConsumerCertLocation
}

func (k *kafka) kfkSetPassphrase(pp string) {
	if pp != "" {
		k.kafkaPassphrase = pp
	}
}
func (k *kafka) KFKGetPassphrase() string {
	return k.kafkaPassphrase
}

func (k *kafka) kfkSetBrokerCertLocation(kl string) {
	if kl != "" {
		k.kafkaBrokerCertLocation = kl
	}
}
func (k *kafka) KFKGetBrokerCertLocation() string {
	return k.kafkaBrokerCertLocation
}

// Config wrap all configs
type Config struct {
	*Global
	*Grpc
	*JwtValidation
	*JWTRequestConfig
	ClientTLSConfig *tls.Config
	ServerTLSConfig *tls.Config
	*observability
	*kafka
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

	// grpc := NewGrpc()
	grpc := NewGRPC()

	jwtValidation := NewJwtValidation()

	c := &Config{
		Global:           g,
		Grpc:             grpc,
		JwtValidation:    jwtValidation,
		JWTRequestConfig: NewJWTRequestConfig(),
		observability:    NewObservability(),
		kafka:            NewKafka(),
	}

	return c
}

// Global may concern any configs
type Global struct {
	golangEnv   string
	serviceName string
	jwtoken     string
}

func NewGlobal(golangEnv, svcName string) *Global {
	return &Global{
		golangEnv:   golangEnv,
		serviceName: svcName}
}

func (g *Global) GlbGetenv() string {
	return g.golangEnv
}

func (g *Global) GlbSetJWToken(tkn string) {
	g.jwtoken = tkn
}

func (g *Global) GlbGetJWToken() string {
	return g.jwtoken
}

func (g *Global) GlbSetServiceName(srvName string) {
	if srvName != "" {
		g.serviceName = srvName
	}
}

func (g *Global) GlbGetServiceName() string {
	return g.serviceName
}

// JwtValidation
type JwtValidation struct {
	jwtValidationAddress string
	jwtValidationPort    string
}

func NewJwtValidation() *JwtValidation {
	jv := &JwtValidation{}

	jv.jwtValidationAddress = jwtValidationAddressDefault
	jv.jwtValidationPort = jwtValidationPortDefault

	return jv
}

func (jv *JwtValidation) JWTVSetAddress(addr string) {
	if addr != "" {
		jv.jwtValidationAddress = addr
	}
}

func (jv *JwtValidation) JWTVGetAddress() string {
	return jv.jwtValidationAddress
}

func (jv *JwtValidation) JWTVSetPort(port string) {
	if port != "" {
		jv.jwtValidationPort = port
	}
}

func (jv *JwtValidation) JWTVGetPort() string {
	return jv.jwtValidationPort
}

// Grpc configs
type Grpc struct {
	address   string
	port      string
	pathToTLS string
}

func NewGRPC() *Grpc {
	h := &Grpc{}

	// set default
	h.address = grpcAddressDefault
	h.port = grpcPortDefault

	return h
}

func (h *Grpc) GRPCSetAddress(addr string) {
	if addr != "" {
		h.address = addr
	}
}

func (h *Grpc) GRPCGetAddress() string {
	return h.address
}

func (h *Grpc) GRPCSetPort(port string) {
	if port != "" {
		h.port = port
	}
}

func (h *Grpc) GRPCGetPort() string {
	return h.port
}

func (h *Grpc) GRPCGetGRPCFormatedURL() string {
	return fmt.Sprintf("http://%v:%v", h.address, h.port)
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
	r := &JWTRequestConfig{}
	r.authSvcURL = authSvcURLDefault
	r.authSvcPath = authSvcPathDefault
	r.authSvcTokenEndpoint = authSvcTokenEndpointDefault
	r.codeVerifier = codeVerifierDefault
	r.serviceKeyID = serviceKeyIDDefault
	r.serviceSecretkey = serviceSecretkeyDefault
	r.scope = scopeDefault
	return r
}

func (r *JWTRequestConfig) JwtSetAuthSvcURL(v string) {
	if v != "" {
		r.authSvcURL = v
	}
}
func (r *JWTRequestConfig) JwtGetAuthSvcURL() string {
	return r.authSvcURL
}

func (r *JWTRequestConfig) JwtSetAuthSvcPath(v string) {
	if v != "" {
		r.authSvcPath = v
	}
}
func (r *JWTRequestConfig) JwtGetAuthSvcPath() string {
	return r.authSvcPath
}

func (r *JWTRequestConfig) JwtSetAuthSvcTokenEndpoint(v string) {
	if v != "" {
		r.authSvcTokenEndpoint = v
	}
}
func (r *JWTRequestConfig) JwtGetAuthSvcTokenEndpoint() string {
	return r.authSvcTokenEndpoint
}

func (r *JWTRequestConfig) JwtSetCodeVerifier(v string) {
	if v != "" {
		r.codeVerifier = v
	}
}
func (r *JWTRequestConfig) JwtGetCodeVerifier() string {
	return r.codeVerifier
}

func (r *JWTRequestConfig) JwtSetServiceKeyID(v string) {
	if v != "" {
		r.serviceKeyID = v
	}
}
func (r *JWTRequestConfig) JwtGetServiceKeyID() string {
	return r.serviceKeyID
}

func (r *JWTRequestConfig) JwtSetServiceSecretKey(v string) {
	if v != "" {
		r.serviceSecretkey = v
	}
}
func (r *JWTRequestConfig) JwtGetServiceSecretKey() string {
	return r.serviceSecretkey
}

func (r *JWTRequestConfig) JwtSetScope(v string) {
	if v != "" {
		r.scope = v
	}
}
func (r *JWTRequestConfig) JwtGetScope() string {
	return r.scope
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
