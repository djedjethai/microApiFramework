package loader

import (
	"os"
)

func setConfigs() Config {

	golangEnv := os.Getenv("GOENV")
	srvName := os.Getenv("SERVICE_NAME")
	c := NewConfig(golangEnv, srvName)
	svcAddr := os.Getenv("SERVICE_ADDRESS")
	c.GlbSetSvcAddress(svcAddr)
	svcPort := os.Getenv("SERVICE_PORT")
	c.GlbSetSvcPort(svcPort)

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

	return c
}
