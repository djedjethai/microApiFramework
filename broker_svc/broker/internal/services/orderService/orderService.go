package orderService

import (
	"context"
	"fmt"
	pb "gitlab.com/grpasr/asonrythme/broker_svc/broker/api/v1/order"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/dto/orderDTO"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
)

// got a grpcHandle, like the service handle the conversion to grpc then the grpcHandle send the message
// got an kafkaHandle, same...
// depends of the request, run async(kafka) or sync(grpc) req

// go generate ./... to build mock
//
//go:generate mockgen -destination=../../mocks/services/mockOrderService.go -package=services gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/orderService IOrderService
type IOrderService interface {
	OrderCreateService(order orderDTO.Order) e.IError
	OrderGetService()
}

type orderService struct {
	grpcClient pb.OrderManagementClient
}

func NewOrderService(grpcClt pb.OrderManagementClient) IOrderService {
	return &orderService{grpcClt}
}

func (o *orderService) OrderCreateService(order orderDTO.Order) e.IError {
	// move the order to grpc or kfk
	ctx := context.Background()

	pbOrder := &pb.Order{}

	// Map the fields from dto.Order to pb.Order
	pbOrder.Id = order.ID
	pbOrder.ShippingAddress = order.ShippingAddress

	// Convert and add LineItems
	for _, dtoLineItem := range order.Items {
		pbLineItem := &pb.LineItem{
			ItemCode: dtoLineItem.ItemCode,
			Quantity: float32(dtoLineItem.Quantity), // Convert to float32 if needed
		}
		pbOrder.LineItem = append(pbOrder.LineItem, pbLineItem)
	}

	// NOTE create a span to trace this req.
	// TODO if exist the span should come be extract from the http req
	ctx, span := obs.Tracing.SPNGetFromCTX(ctx, "brokerSvc_createOrder")
	defer span.End()

	// send the grpc request
	res, err := o.grpcClient.CreateOrder(ctx, pbOrder)
	if err != nil {
		// TODO add err to the span
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("failed CreateOrder")
		return e.NewCustomHTTPStatus(e.StatusBadRequest)
	}

	fmt.Println("the grpc response: ", res)

	ce := e.NewCustomHTTPStatus(e.StatusOK)
	pl := make(map[string]interface{})
	pl["order_id"] = res.Value
	ce.SetPayload(pl)

	return ce
}

func (o *orderService) OrderGetService() {
	fmt.Println("hit GetOrderService....")
}
