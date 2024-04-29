package broker

import (
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
)

func setConfigs() (*config.Config, error) {
	return config.SetConfigs()
}
