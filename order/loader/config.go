package loader

import (
	"fmt"
)

const (
	// global
	GOENVDefault   string = "localhost"
	svcNameDefault string = "order"
	// svcURLDefault         = "http://localhost:8080"
	svcAddressDefault = "localhost"
	svcPortDefault    = "50001"

	// JWTrequestConfigs
	authSvcURLDefault           string = "http://localhost:9096/v1" // all varEnv
	authSvcPathDefault                 = "apiauth"
	authSvcTokenEndpointDefault        = "http://localhost:9096/v1/oauth/token"
	codeVerifierDefault                = "exampleCodeVerifier"
	serviceKeyIDDefault                = "order"
	serviceSecretkeyDefault            = "orderSecret"
	scopeDefault                       = "read, openid"
)

// Config wrap all configs
type Config struct {
	*Global
	*JWTRequestConfig
}

func NewConfig(goEnv string, serviceName ...string) Config {

	svcName := svcNameDefault
	if len(serviceName) > 0 && serviceName[0] != "" {
		svcName = serviceName[0]
	}
	env := GOENVDefault
	if len(goEnv) > 0 {
		env = goEnv
	}

	g := NewGlobal(env, svcName)

	c := Config{
		Global:           g,
		JWTRequestConfig: NewJWTRequestConfig(),
	}

	return c
}

// Global may concern any configs
type Global struct {
	GOENV   string
	svcName string
	// svcURL  string
	svcAddress string
	svcPort    string
}

func NewGlobal(goEnv, svcName string) *Global {
	return &Global{
		GOENV:      goEnv,
		svcName:    svcName,
		svcAddress: svcAddressDefault,
		svcPort:    svcPortDefault}
}

func (g *Global) GlbGetenv() string {
	return g.GOENV
}

func (g *Global) GlbSetSvcName(svcName string) {
	if svcName != "" {
		g.svcName = svcName
	}
}

func (g *Global) GlbGetSvcName() string {
	return g.svcName
}

func (g *Global) GlbSetSvcAddress(svcAddr string) {
	if svcAddr != "" {
		g.svcAddress = svcAddr
	}
}

func (g *Global) GlbSetSvcPort(svcPort string) {
	if svcPort != "" {
		g.svcPort = svcPort
	}
}

func (g *Global) GlbGetHTTPSvcURL() string {
	return fmt.Sprintf("http://%v:%v", g.svcAddress, g.svcPort)
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
