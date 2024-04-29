package loader

import (
	"context"
	"fmt"
	"gitlab.com/grpasr/common/apiserver"
	"gitlab.com/grpasr/common/configloader"
	e "gitlab.com/grpasr/common/errors/json"
	"gitlab.com/grpasr/common/restclient"
	"log"
	"os"
	"path/filepath"
	"time"
)

var jwtoken string

func init() {

	conf := setConfigs()

	// get jwtoken from auth_svc
	apiServerAuth := apiserver.NewAPIserverAuth(
		conf.JwtGetAuthSvcURL(),           // "http://localhost:9096/v1", // all varEnv
		conf.JwtGetAuthSvcPath(),          // "apiauth",
		conf.GlbGetHTTPSvcURL(),           // "http://localhost:8080",
		conf.JwtGetAuthSvcTokenEndpoint(), // "http://localhost:9096/v1/oauth/token",
		conf.JwtGetCodeVerifier(),         //"exampleCodeVerifier",
		conf.JwtGetServiceKeyID(),         // "brokerSvc",
		conf.JwtGetServiceSecretKey(),     // "brokerSvcSecret",
		conf.JwtGetScope(),                // "read, openid",
	)

	err := apiServerAuth.Run(context.TODO(), int8(3), int8(5))
	if err != nil {
		log.Fatal("Order svc error authentication: ", err)
	}

	jwtoken = apiServerAuth.GetToken()
	fmt.Println("broker_svc/loader see jwtoken: ", jwtoken)
}

func Run() {

	goenv := os.Getenv("GOENV")

	c, err := configloader.NewConfig(goenv, configloader.TLoaderConfig)
	if err != nil {
		fmt.Println("in err: ", err)
		log.Fatal(err)
	}

	configLoaderPath := os.Getenv("CONFIG_LOADER_PATH")
	configLoaderFilename := os.Getenv("CONFIG_LOADER_FILENAME")

	c.LDRLoadConfigs(configLoaderFilename, "yaml", configLoaderPath)
	// c.LDRLoadConfigs("loaderConfig", "yaml", "../loader/configs/")

	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// add Bearer token to configs
	ad := restclient.AuthData{
		AuthType: restclient.Bearer,
		Token:    jwtoken,
		// Token: "",
		// Token: "eyJhbGciOiJIUzI1NiIsImtpZCI6InRoZUtleUlEIiwidHlwIjoiSldUIn0.eyJhdWQiOiJvcmRlciIsImV4cCI6MTcwMjI3ODI4OCwic3ViIjoib3JkZXIiLCJvcGVuaWRJbmZvIjp7ImFnZSI6MzUsImNpdHkiOiJMb25kb24iLCJuYW1lIjoiUm9iZXJ0Iiwicm9sZSI6IkFQSXNlcnZlciIsInNjb3BlIjoicmVhZCwgb3BlbmlkIn19.z7yz26jTguXsBYeWts6k9MImp11vRHW0_E9", // expired
	}

	fmt.Println("See the endpoint to make get the configs: ", c.LDRConfig().GetServiceEndpointFormatedURL())

	// the url format must be "http://path:port"
	conf := restclient.NewConfig(fmt.Sprintf("http://%s", c.LDRConfig().GetServiceEndpointFormatedURL()), ad)

	restSvc, err := restclient.NewRestService(conf, "multipart/form-data")
	if err != nil {
		fmt.Println("Error from NewRestService: ", err)
	}

	// set context to time up
	totReq := len(c.LDRGetGrpcTypes()) + len(c.LDRGetConfigsFiles())
	reqR := c.LDRConfig().LcfgGetReqRetry()
	reqRetDel := c.LDRConfig().LcfgGetDelayBetweenReqRetry()
	totTime := int((reqR*reqR)*reqRetDel) * totReq
	totDelay, _ := time.ParseDuration(fmt.Sprintf("%vs", totTime))

	ctx, cancel := context.WithTimeout(context.Background(), totDelay)
	defer cancel()

	// download the grpcSchemas
	var errIErr e.IError
	for _, grpcPackage := range c.LDRGetGrpcTypes() {
		grpcStorageDir := filepath.Join(
			currentDir,
			c.LDRConfig().LcfgGetStoragePathGrpc(),
			c.LDRConfig().LcfgGetVersion(),
			grpcPackage)
		fmt.Println("broker_svc loader, the grpcStorageDir.....: ", grpcStorageDir)

		downloadURL := filepath.Join(
			c.LDRConfig().LcfgGetDownloadPathGrpc(),
			c.LDRConfig().LcfgGetVersion(),
			grpcPackage)
		fmt.Println("broker_svc loader, the downloadURL.....: ", downloadURL)

		errIErr = restSvc.HandleRetryRequest(
			ctx,
			restclient.NewRequest("GET", downloadURL, nil),
			nil,
			reqR,
			reqRetDel,
			restclient.THandleMultipartWriter,
			grpcStorageDir,
		)
		if errIErr != nil {
			// fmt.Println("errror from loader: ", errIErr.GetCode())
			switch errIErr.GetCode() {
			case 401:
				log.Println("loading GRPC types, jwtoken expired: ", errIErr.Error())
			case 403:
				log.Println("loading GRPC types, jwtoken invalid: ", errIErr.Error())
			}
		}
	}

	// download the configs files
	for _, configFile := range c.LDRGetConfigsFiles() {
		configStorageDir := filepath.Join(
			currentDir,
			c.LDRConfig().LcfgGetStoragePathConfigs(),
			c.LDRConfig().LcfgGetVersion(),
			configFile)
		downloadURL := filepath.Join(
			c.LDRConfig().LcfgGetDownloadPathConfigs(),
			c.LDRConfig().LcfgGetVersion(),
			configFile)

		errIErr = restSvc.HandleRetryRequest(
			ctx,
			restclient.NewRequest("GET", downloadURL, nil),
			nil,
			reqR,
			reqRetDel,
			restclient.THandleMultipartWriter,
			configStorageDir,
		)
		if errIErr != nil {
			// fmt.Println("errror from loader: ", errIErr.GetCode())
			switch errIErr.GetCode() {
			case 401:
				log.Println("loading config files, jwtoken expired: ", errIErr.Error())
			case 403:
				log.Println("loading config files, jwtoken invalid: ", errIErr.Error())
			}
		}
	}
}
