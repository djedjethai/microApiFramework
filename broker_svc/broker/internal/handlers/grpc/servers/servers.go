package servers

import (
	"fmt"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services"
	obs "gitlab.com/grpasr/common/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

// GRPCServer is an handle to all grpc servers
type GRPCServers struct {
	configs         *config.Config
	brokerjwtServer *grpc.Server
}

func NewGRPCServers(svc services.GrpcServices, configs *config.Config) (*GRPCServers, error) {
	def := &GRPCServers{}

	// set tls options
	opts := []grpc.ServerOption{}
	creds := credentials.NewTLS(configs.ServerTLSConfig)
	opts = append(opts, grpc.Creds(creds))

	brokerjwtServer, err := NewGrpcBrokerjwtServer(svc, opts...)
	if err != nil {
		return def, err
	}

	return &GRPCServers{
		configs:         configs,
		brokerjwtServer: brokerjwtServer,
	}, nil
}

func (gs *GRPCServers) GRPCServersListen() {

	// NOTE start all grpc servers here
	go grpcListen(gs.brokerjwtServer, gs.configs.GRPCGetJwtValidationPort())
}

func grpcListen(serv *grpc.Server, grpcPort string) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcPort))
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHFatal()).
			Err(err).
			Msg("err creating grpc listener")
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHInfo()).
		Str("GRPC listen on port: ", grpcPort).
		Send()
	err = serv.Serve(lis)
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHFatal()).
			Err(err).
			Msg("err grpc server listen")
	}
}
