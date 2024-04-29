package order

import (
	"order/order/internal/config"
)

func setConfigs() (*config.Config, error) {
	return config.SetConfigs()
}
