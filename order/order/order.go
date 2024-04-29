package order

import (
	"context"
	"fmt"
	kafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"gitlab.com/grpasr/common/apiserver"
	obs "gitlab.com/grpasr/common/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
	"net"
	pbb "order/order/api/v1/brokerjwt"
	pb "order/order/api/v1/order"
	"order/order/internal/config"
	"order/order/internal/service"
	"order/order/internal/types"
)

const (
	grpcPort   = "50001"
	nullOffset = -1
)

var (
	jwtoken string
	conf    *config.Config
	kfk     *Kafka
)

// init authenticate the service with the auth_svc and get the service's jwt_token
func init() {

	var err error
	conf, err = setConfigs()
	if err != nil {
		log.Fatal("Error setting configurations")
	}

	// get jwtoken from auth_svc
	apiServerAuth := apiserver.NewAPIserverAuth(
		conf.JwtGetAuthSvcURL(),           // "http://localhost:9096/v1", // all varEnv
		conf.JwtGetAuthSvcPath(),          // "apiauth",
		conf.GRPCGetGRPCFormatedURL(),     // "http://localhost:50001",
		conf.JwtGetAuthSvcTokenEndpoint(), // "http://localhost:9096/v1/oauth/token",
		conf.JwtGetCodeVerifier(),         //"exampleCodeVerifier",
		conf.JwtGetServiceKeyID(),         // "order",
		conf.JwtGetServiceSecretKey(),     // "orderSecret",
		conf.JwtGetScope(),                // "read, openid",
	)

	err = apiServerAuth.Run(context.TODO(), int8(3), int8(8))
	if err != nil {
		log.Fatal("Order svc error authentication: ", err)
	}

	jwtoken = apiServerAuth.GetToken()

	fmt.Println("see the jwtoken: ", jwtoken)
}

func Run() {

	// set kafka
	k := NewKafka()
	k.setProducer()
	kfk = k

	// set observability
	obs.SetObservabilityFacade()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// the storage
	orderSvc := service.NewOrderSvc()

	// NOTE test brokerjwt validation ========================
	bjc := NewBrokerJWTClient()
	err := bjc.setBrokerClient()
	if err != nil {
		log.Fatal(err)
	}

	token := &pbb.Jwtoken{
		Jwt: jwtoken,
		// Jwt: "eyJhbGciOiJIUzI1NiIsImtpZCI6InRoZUtleUlEIiwidHlwIjoiSldUIn0.eyJhdWQiOiJvcmRlciIsImV4cCI6MTcwMjI3ODI4OCwic3ViIjoib3JkZXIiLCJvcGVuaWRJbmZvIjp7ImFnZSI6MzUsImNpdHkiOiJMb25kb24iLCJuYW1lIjoiUm9iZXJ0Iiwicm9sZSI6IkFQSXNlcnZlciIsInNjb3BlIjoicmVhZCwgb3BlbmlkIn19.z7yz26jTguXsBYeWts6k9MImp11vRHW0_E9", // expired
		// Jwt: "eyJhbGciOiJIUzI1NiIsImtpZCI6InRoZUtleUlEIiwidHlwIjoiSldUIn0.eyJhdWQiOiJvcmRlciIsImV4cCI6MTcwNDg3MTY3MSwic3ViIjoib3JkZXIiLCJvcGVuaWRJbmZvIjp7ImFnZSI6MzUsImNpdHkiOiJMb25kb24iLCJuYW1lIjoiUm9iZXJ0Iiwicm9sZSI6IkFQSXNlcnZlciIsInNjb3BlIjoicmVhZCwgb3BlbmlkIn19.vJ_MWk6px5iluyjf6YkrMuybbAw", // invalid
	}

	// NOTE set the tracing
	if conf.GlbGetenv() != "localhost" {
		tp, err := obs.Tracing.SetupTracing(
			ctx,
			conf.ClientTLSConfig,
			conf.OBSGetSampling(),
			conf.GlbGetServiceName(),
			conf.OBSGetCollectorEndpoint(),
			conf.GlbGetenv())
		if err != nil {
			panic(err)
		}
		defer tp.Shutdown(ctx)

		mp, err := obs.Metrics.SetupMetrics(
			ctx,
			conf.ClientTLSConfig,
			conf.OBSGetScratchDelay(),
			conf.GlbGetServiceName(),
			conf.OBSGetCollectorEndpoint(),
			conf.GlbGetenv())
		if err != nil {
			panic(err)
		}
		defer mp.Shutdown(ctx)

	}

	// TODO if err connection(as invalid tls for eg) response is an invalid mem addr
	response, err := bjc.getBrokerClient().IsJwtokenOK(context.TODO(), token)
	if err != nil {
		fmt.Println("Error when validate brokerjwt")
	}
	fmt.Println("See the brokerjwt response: ", response.IsOk)
	fmt.Println("See the brokerjwt response: ", response.ResponseCode)
	fmt.Println("See the brokerjwt overall response: ", response)
	// end test brokerjwt validation ===========================

	// http.HandleFunc("/orders", createOrders(&orderStr))
	// http.HandleFunc("/order", getOrders(&orderStr))

	fmt.Println("order Listen on grpc port 50001... !!")
	// log.Fatal(http.ListenAndServe(":8080", nil))
	grpcListen(&orderSvc)
}

type Kafka struct {
	Producer *kafka.Producer
}

func NewKafka() *Kafka {
	return &Kafka{}
}

func (k *Kafka) setProducer() {

	fmt.Println("see KFK URL: ", conf.KFKGetURL())
	fmt.Println("see ProducerKeyLocation: ", conf.KFKGetProducerKeyLocation())
	fmt.Println("see ProducerCertLocation: ", conf.KFKGetProducerCertLocation())
	fmt.Println("see BrokerCertLocation: ", conf.KFKGetBrokerCertLocation())

	// Create a new producer instance
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":        conf.KFKGetURL(),
		"security.protocol":        "ssl",
		"ssl.key.location":         conf.KFKGetProducerKeyLocation(),
		"ssl.key.password":         "datahub",
		"ssl.certificate.location": conf.KFKGetProducerCertLocation(),
		// selfSigned cert the cert is the ca as well
		"ssl.ca.location":                     conf.KFKGetBrokerCertLocation(),
		"enable.ssl.certificate.verification": false,
		"debug":                               "security,broker",
	})
	if err != nil {
		log.Fatal(err)
	}

	k.Producer = p
}

func (k *Kafka) produce(msg proto.Message, topic string) (int64, error) {
	kafkaChan := make(chan kafka.Event)
	defer close(kafkaChan)

	fmt.Println("produce 1")

	// Serialize the protobuf message to bytes
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return 1, err
	}

	// fmt.Println("produce 2: ", msgBytes)

	if err := k.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny},
		Value: msgBytes,
		// Value: msgBytes,
	}, kafkaChan); err != nil {
		return nullOffset, err
	}

	fmt.Println("produce 3")

	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		log.Println("message sent: ", string(ev.Value))
		return int64(ev.TopicPartition.Offset), nil
	case kafka.Error:
		return nullOffset, err
	}
	return nullOffset, nil

}

func grpcListen(orderSvc *service.OrderSvc) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcPort))
	if err != nil {
		log.Fatal("err creating the listener: ", err)
	}

	// set tls options
	opts := []grpc.ServerOption{}
	// opts := make([]grpc.ServerOption, 2)
	creds := credentials.NewTLS(conf.ServerTLSConfig)
	opts = append(opts, grpc.Creds(creds))

	if conf.GlbGetenv() != "localhost" {
		// opts = append(opts, grpc.UnaryInterceptor(traceInterceptor))
		opts = append(opts, grpc.UnaryInterceptor(obs.Tracing.GRPCTraceInterceptorServer))
	}

	serv, err := NewGrpcServer(orderSvc, opts...)
	if err != nil {
		log.Fatal("err creating the server: ", err)
	}

	log.Println("order grpc server is listening on port: ", grpcPort)
	err = serv.Serve(lis)
	if err != nil {
		log.Println("err server listen: ", err)
	}
}

// client for brokerjwt
type BrokerJWTClient struct {
	client pbb.JwtokenManagementClient
}

func NewBrokerJWTClient() *BrokerJWTClient {
	return &BrokerJWTClient{}
}

func (bc *BrokerJWTClient) setBrokerClient() error {
	ctx := context.Background()

	// set TLS
	creds := credentials.NewTLS(conf.ClientTLSConfig)

	conn, err := grpc.DialContext(
		ctx,
		// "127.0.0.1:50002", // localhost
		// "broker_svc:50002", // dev
		fmt.Sprintf("%v:%v", conf.JWTVGetAddress(), conf.JWTVGetPort()),
		grpc.WithTransportCredentials(creds),
		// grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	fmt.Println("grpc order client is ready")
	client := pbb.NewJwtokenManagementClient(conn)
	bc.client = client
	return nil
}

func (bc *BrokerJWTClient) getBrokerClient() pbb.JwtokenManagementClient {
	return bc.client
}

// server
type Server struct {
	pb.UnimplementedOrderManagementServer
	orderSvc *service.OrderSvc
}

func NewGrpcServer(ordSvc *service.OrderSvc, opts ...grpc.ServerOption) (*grpc.Server, error) {
	gsrv := grpc.NewServer(opts...)

	srv := &Server{
		orderSvc: ordSvc,
	}

	pb.RegisterOrderManagementServer(gsrv, srv)

	return gsrv, nil
}

func (s *Server) CreateOrder(ctx context.Context, order *pb.Order) (*wrapperspb.StringValue, error) {

	fmt.Println("get the order id: ", order.Id)
	fmt.Println("get the order shippingAddress: ", order.ShippingAddress)
	// fmt.Println("get the order : ", order.Id)

	// NOTE tracing staff
	if conf.GlbGetenv() != "localhost" {
		_, span := obs.Tracing.SPNGetFromCTX(
			ctx,
			"orderSvc-createOrder",
			obs.Tracing.TAString("service", "order"),
			obs.Tracing.TAString("function", "CreateOrder"),
		)
		defer span.End()
	}

	// kafka send message
	offset, err := kfk.produce(order, "order")
	if err != nil {
		fmt.Println("Error producing message to kafka: ", err)
	}
	fmt.Println("Message produced to kafka, see the offset: ", offset)

	o := types.Order{}
	// create an uid
	newUUID := uuid.New()
	uuidString := newUUID.String()
	o.ID = uuidString
	o.ShippingAddress = order.ShippingAddress

	o.Items = []types.LineItem{}
	for _, item := range order.LineItem {
		fmt.Println("Item: ", item.ItemCode)
		fmt.Println("Item qtt: ", item.Quantity)
		li := types.LineItem{
			ItemCode: item.ItemCode,
			Quantity: int(item.Quantity),
		}
		o.Items = append(o.Items, li)
	}

	// save the received data
	s.orderSvc.PlaceOrder(o)

	return &wrapperspb.StringValue{Value: uuidString}, nil
}
