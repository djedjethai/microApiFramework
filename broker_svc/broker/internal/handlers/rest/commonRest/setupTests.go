package commonRest

import (
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"testing"
)

type SetupTests struct {
	router *mux.Router
	ctrl   *gomock.Controller
	t      *testing.T
}

func NewSetupTests() *SetupTests {
	t := &testing.T{}

	ctrl := gomock.NewController(t)
	router := mux.NewRouter()

	return &SetupTests{router, ctrl, t}
}

func (st *SetupTests) GetRouter() *mux.Router {
	return st.router
}

func (st *SetupTests) GetCtrl() *gomock.Controller {
	return st.ctrl
}

func (st *SetupTests) Teardown() {
	// Perform any teardown here
	st.ctrl.Finish()
	st.router = nil
}

// var Router *mux.Router
// var Ctrl *gomock.Controller

// Order
// var mockOrderService *services.MockIOrderService
// var orderRestT *orderRest.OrderRest

// func TestMain(m *testing.M) {
// 	// Initialize your test context
// 	testContext := setup()
//
// 	// Run tests
// 	exitCode := m.Run()
//
// 	// Teardown
// 	teardown(testContext)
//
// 	// Exit with the appropriate code
// 	os.Exit(exitCode)
//
// }

// func Setup() *testing.T {
// 	t := &testing.T{}
//
// 	ctrl = gomock.NewController(t)
// 	router = mux.NewRouter()
//
// 	// create the mockOrderSvc
// 	mockOrderService = services.NewMockIOrderService(ctrl)
// 	orderRestT = orderRest.NewOrderRest(router, mockOrderService)
//
// 	// NOTE create upcomming handlers
//
// 	return t
// }
