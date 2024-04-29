package clients

import (
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
)

// const grpcPort = "50001"

// GRPCHandler is an handle to all grpc clients
type GRPCClients struct {
	*OrderGrpc
}

func NewGRPCClients(config *config.Config) (GRPCClients, error) {
	def := GRPCClients{}

	// orderClient
	orderGRPC, err := NewOrderGrpc(config)
	if err != nil {
		return def, err
	}

	grpcHandler := GRPCClients{
		OrderGrpc: orderGRPC,
	}
	return grpcHandler, nil
}
