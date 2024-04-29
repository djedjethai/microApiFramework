package main

import (
	"github.com/djedjethai/go-oauth2-openid/models"
	mongo "github.com/djedjethai/mongo-openid"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/config"
)

func setConfigs() (*config.Config, error) {
	return config.SetConfigs()
}

func registerServices(clientStore *mongo.ClientStore, conf *config.Config) {
	svcs := conf.SVCGetServices()

	// register the frontend
	clientStore.Create(&models.Client{
		ID:     svcs["frontend"].SVCGetID(),
		Secret: svcs["frontend"].SVCGetSecret(),
		Domain: svcs["frontend"].SVCGetDomain(),
		UserID: "frontend",
	})

	// TODO seems that if credentials are changed, new credentials are not updated...
	// TODO see what the difference between ID and UserID
	// (are the same with configurations right now)...
	clientStore.Create(&models.Client{
		ID:     svcs["order"].SVCGetID(),
		Secret: svcs["order"].SVCGetSecret(),
		Domain: svcs["order"].SVCGetDomain(),
		UserID: "order",
	})

	// register broker_svc
	clientStore.Create(&models.Client{
		ID:     svcs["broker_svc"].SVCGetID(),
		Secret: svcs["broker_svc"].SVCGetSecret(),
		Domain: svcs["broker_svc"].SVCGetDomain(),
		UserID: "brokerSvc",
	})

	// register registry_svc
	clientStore.Create(&models.Client{
		ID:     svcs["registry_svc"].SVCGetID(),
		Secret: svcs["registry_svc"].SVCGetSecret(),
		Domain: svcs["registry_svc"].SVCGetDomain(),
		UserID: "registrySvc",
	})
}
