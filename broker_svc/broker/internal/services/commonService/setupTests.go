package commonService

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

func RunBackgroundServer(serv *grpc.Server, grpcPort string) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcPort))
	if err != nil {
		log.Fatal("err creating the listener: ", err)
	}

	log.Println("order grpc server is listening on port: ", grpcPort)
	err = serv.Serve(lis)
	if err != nil {
		log.Println("err server listen: ", err)
	}
}
