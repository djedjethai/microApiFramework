package servers

import (
	"context"
	"fmt"
	pb "gitlab.com/grpasr/asonrythme/broker_svc/broker/api/v1/brokerjwt"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/jwtokenService"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

type BrokerjwtServer struct {
	pb.UnimplementedJwtokenManagementServer
	jwtSvc *jwtokenService.JWTokenService
	// TODO add the service which handle the jwt auth
}

// func NewGrpcServer(ordSvc *service.OrderSvc, opt ...grpc.ServerOption) (*grpc.Server, error) {
func NewGrpcBrokerjwtServer(svc services.GrpcServices, opts ...grpc.ServerOption) (*grpc.Server, error) {
	// func NewGrpcBrokerjwtServer(opt ...grpc.ServerOption) (*grpc.Server, error) {
	gsrv := grpc.NewServer(opts...)

	bsrv := &BrokerjwtServer{
		jwtSvc: svc.JWTokenService,
	}

	pb.RegisterJwtokenManagementServer(gsrv, bsrv)

	return gsrv, nil
}

func (bs *BrokerjwtServer) IsJwtokenOK(ctx context.Context, jwt *pb.Jwtoken) (*pb.IsJwtokenOKResponse, error) {
	// TODO add the service brokerjwt which will validate the token

	fmt.Println("In BrokerjwtServer - IsJwtokenOK - see the token: ", jwt.Jwt)

	resp := &pb.IsJwtokenOKResponse{}

	// validate the token
	infos, err := bs.jwtSvc.JWTokenIsValidToken(context.TODO(), jwt.Jwt)
	if err != nil {
		resp.IsOk = false
		resp.ResponseCode = int32(err.GetCode())
	} else {
		resp.IsOk = true
		resp.ResponseCode = 200
		resp.Role = infos["role"]
		resp.Svc = infos["svc"]
		resp.Scope = infos["scope"]
	}

	return resp, nil
}
