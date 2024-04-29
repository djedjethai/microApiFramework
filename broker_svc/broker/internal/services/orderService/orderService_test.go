package orderService

import (
	"context"
	"fmt"
	pb "gitlab.com/grpasr/asonrythme/broker_svc/broker/api/v1/order"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/dto/orderDTO"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/commonService"
	obs "gitlab.com/grpasr/common/observability"
	"gitlab.com/grpasr/common/tests"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

// var wg sync.WaitGroup
var grpcPort = "50001"
var grpcAddress = "127.0.0.1"
var client pb.OrderManagementClient
var serv *grpc.Server
var orderServiceT IOrderService

func TestMain(m *testing.M) {
	// Initialize your test context
	setupTests()

	// Run tests
	exitCode := m.Run()

	// Teardown
	teardown()

	// Exit with the appropriate code
	os.Exit(exitCode)

}

func Test_create_an_order_fail_if_invalid_grpc_connection(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)
	// mock the service input
	it := orderDTO.LineItem{
		ItemCode: "theitemcode",
		Quantity: 12,
	}
	order := orderDTO.Order{
		ID:              "123",
		Items:           []orderDTO.LineItem{it},
		ShippingAddress: "the shipping address",
	}

	ce := orderServiceT.OrderCreateService(order)

	// Assert
	tests.MaybeFail("http_status", tests.Expect(ce.GetCode(), http.StatusBadRequest))
	tests.MaybeFail("http_status", tests.Expect(len(ce.GetPayload()), 0))
}

func Test_create_an_order_is_successful(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	// run the background server
	go commonService.RunBackgroundServer(serv, grpcPort)

	// give some time to the server to be up running
	time.Sleep(2 * time.Second)

	// mock the service input
	it := orderDTO.LineItem{
		ItemCode: "theitemcode",
		Quantity: 12,
	}
	order := orderDTO.Order{
		ID:              "123",
		Items:           []orderDTO.LineItem{it},
		ShippingAddress: "the shipping address",
	}

	ce := orderServiceT.OrderCreateService(order)

	// Assert
	tests.MaybeFail("http_status", tests.Expect(ce.GetCode(), http.StatusOK))
	tests.MaybeFail("http_status", tests.Expect(len(ce.GetPayload()), 1))
	tests.MaybeFail("http_status", tests.Expect(ce.GetPayload()["order_id"], order.ID))
}

// MOCK THE SERVER
type OrderServer struct {
	pb.UnimplementedOrderManagementServer
}

func (s *OrderServer) CreateOrder(ctx context.Context, order *pb.Order) (*wrapperspb.StringValue, error) {

	// fmt.Println("get the order id: ", order.Id)
	// fmt.Println("get the order shippingAddress: ", order.ShippingAddress)
	// fmt.Println("get the order : ", order.Id)

	// TODO add what ever needed logic

	return &wrapperspb.StringValue{Value: order.Id}, nil
}

func setupTests() {

	obs.SetObservabilityFacade("orderTest")

	err := setClient()
	if err != nil {
		log.Fatal("err creating the client: ", err)
	}

	serv, err = MockGrpcOrderServer()
	if err != nil {
		log.Fatal("err creating the server: ", err)
	}

	orderServiceT = NewOrderService(client)
}

func teardown() {

}

// SET THE CLIENT
func setClient() error {
	ctx := context.Background()

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:%s",
			grpcAddress,
			grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil
	}
	client = pb.NewOrderManagementClient(conn)

	return nil
}

// SET THE SERVER
func MockGrpcOrderServer() (*grpc.Server, error) {
	opts := []grpc.ServerOption{}

	// func NewGrpcBrokerjwtServer(opt ...grpc.ServerOption) (*grpc.Server, error) {
	gsrv := grpc.NewServer(opts...)

	osrv := &OrderServer{
		// jwtSvc: svc.JWTokenService,
	}

	pb.RegisterOrderManagementServer(gsrv, osrv)

	return gsrv, nil
}
