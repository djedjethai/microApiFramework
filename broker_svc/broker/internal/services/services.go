package services

import (
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/grpc/clients"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/jwtokenService"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/orderService"
)

// RestServices Wrap all services needed for the rest handlers
type RestServices struct {
	// *OrderService
	OrderService orderService.IOrderService
	*jwtokenService.JWTokenService
}

func NewRestServices(grpcHandler clients.GRPCClients, conf *config.Config) RestServices {
	svc := RestServices{
		OrderService:   orderService.NewOrderService(grpcHandler.OrderGetClient()),
		JWTokenService: jwtokenService.NewJWTokenService(conf),
	}
	return svc
}

// GrpcServices Wrap all services needed for the grpc handlers
type GrpcServices struct {
	*jwtokenService.JWTokenService
}

func NewGrpcServices(grpcHandler clients.GRPCClients, conf *config.Config) GrpcServices {
	svc := GrpcServices{
		JWTokenService: jwtokenService.NewJWTokenService(conf),
	}
	return svc
}
