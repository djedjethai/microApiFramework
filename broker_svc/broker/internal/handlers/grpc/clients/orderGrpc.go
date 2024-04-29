package clients

import (
	"context"
	"fmt"
	pb "gitlab.com/grpasr/asonrythme/broker_svc/broker/api/v1/order"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	obs "gitlab.com/grpasr/common/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	svcOrder = "order"
)

type OrderGrpc struct {
	client pb.OrderManagementClient
}

func NewOrderGrpc(conf *config.Config) (*OrderGrpc, error) {
	og := &OrderGrpc{}
	err := og.orderSetClient(conf)
	return og, err
}

func (o *OrderGrpc) orderSetClient(conf *config.Config) error {
	ctx := context.Background()

	// set TLS
	creds := credentials.NewTLS(conf.ClientTLSConfig)

	// set clientConnection
	var conn *grpc.ClientConn
	var err error
	orderSvcInfo := conf.SVCSGetServices()[svcOrder]
	if conf.GlbGetenv() != "localhost" {
		traceDialOption := grpc.WithUnaryInterceptor(obs.Tracing.GRPCTraceInterceptorClient)
		conn, err = grpc.DialContext(
			ctx,
			fmt.Sprintf("%s:%s",
				orderSvcInfo.SVCGetAddress(),
				orderSvcInfo.SVCGetPort()),
			grpc.WithTransportCredentials(creds),
			traceDialOption,
			// grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

	} else {
		// conn, err = setConnection(ctx, orderSvcInfo, &creds)
		conn, err = grpc.DialContext(
			ctx,
			fmt.Sprintf("%s:%s",
				orderSvcInfo.SVCGetAddress(),
				orderSvcInfo.SVCGetPort()),
			grpc.WithTransportCredentials(creds),
		)
	}
	if err != nil {
		return err
	}

	obs.Logging.NewLogHandler(obs.Logging.LLHInfo()).
		Msg("grpc order client is ready")
	client := pb.NewOrderManagementClient(conn)
	o.client = client
	return nil
}

func (o *OrderGrpc) OrderGetClient() pb.OrderManagementClient {
	return o.client
}
