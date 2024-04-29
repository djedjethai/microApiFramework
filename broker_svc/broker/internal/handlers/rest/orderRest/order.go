package orderRest

import (
	"fmt"
	"github.com/gorilla/mux"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/dto/orderDTO"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/rest/commonRest"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/services/orderService"
	"net/http"
)

type OrderRest struct {
	orderRouter *mux.Router
	orderSvc    orderService.IOrderService
}

func NewOrderRest(orderRouter *mux.Router, ordSvc orderService.IOrderService) *OrderRest {
	return &OrderRest{orderRouter, ordSvc}
}

func (o *OrderRest) RunOrderRest() {
	o.orderRouter.HandleFunc("/v1/createorder", o.CreateOrder).
		Methods(http.MethodPost).
		Name("CreateOrder")
	o.orderRouter.HandleFunc("/v1/getorder", o.GetOrder).
		Methods(http.MethodPost).
		Name("GetOrder")
	// o.orderRouter.HandleFunc("/{customer_id:[0-9]+}/order", o.PostOrderRequest).
	// 	Methods(http.MethodGet).
	// 	Name("NewAccount")
}

// Order service handlers

func (o *OrderRest) CreateOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("internal/handlers/rest - CreateOrder.........")
	fmt.Fprintln(w)

	li := orderDTO.LineItem{}
	li.ItemCode = "theitemcode"
	li.Quantity = 12

	// create a test Order
	or := orderDTO.Order{}
	or.ID = "123"
	or.Items = []orderDTO.LineItem{li}
	or.ShippingAddress = "the shipping address"

	ce := o.orderSvc.OrderCreateService(or)

	// // if use writeJson
	// headers1 := make(http.Header)
	// headers1.Add("Content-Type", "application/json")

	commonRest.CustomResponseJson(w, ce)
}

func (o *OrderRest) GetOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w)

}

// // Handle POST request for account creation
// func (o *OrderRest) PostOrderRequest(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	customerID := vars["customer_id"]
// 	fmt.Fprintf(w, "Handling POST request to create account for customer %s\n", customerID)
// }
