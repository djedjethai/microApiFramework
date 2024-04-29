package orderRest

import (
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/dto/orderDTO"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/rest/commonRest"
	// "gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/handlers/rest/orderRest"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/mocks/services"
	e "gitlab.com/grpasr/common/errors/json"
	"gitlab.com/grpasr/common/tests"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var st *commonRest.SetupTests

var mockOrderService *services.MockIOrderService
var orderRestT *OrderRest

func TestMain(m *testing.M) {
	// Initialize your test context
	st = commonRest.NewSetupTests()

	mockOrderService = services.NewMockIOrderService(st.GetCtrl())
	orderRestT = NewOrderRest(st.GetRouter(), mockOrderService)

	// Run tests
	exitCode := m.Run()

	// Teardown
	st.Teardown()

	// Exit with the appropriate code
	os.Exit(exitCode)

}

func Test_create_order_create_an_order(t *testing.T) {
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

	// mock the service output
	ce := e.NewCustomHTTPStatus(e.StatusOK)
	pl := make(map[string]interface{})
	pl["order_id"] = "12345"
	ce.SetPayload(pl)

	// ceJSON should be match by the response payload
	ceJSON, err := json.Marshal(ce)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// NOTE right now as the OrderCreateService() input is hard coded
	// order must match it.....
	mockOrderService.EXPECT().OrderCreateService(order).Return(ce)

	st.GetRouter().HandleFunc("/v1/createorder", orderRestT.CreateOrder)
	request, _ := http.NewRequest(http.MethodPost, "/v1/createorder", nil)

	// Act
	recorder := httptest.NewRecorder()
	st.GetRouter().ServeHTTP(recorder, request)

	// Assert
	tests.MaybeFail("http_status", tests.Expect(recorder.Code, http.StatusOK))
	tests.MaybeFail("http_content_type", tests.Expect(
		recorder.Header().Get("Content-Type"), "application/json"))
	tests.MaybeFail("response_body", tests.Expect(
		strings.TrimSpace(recorder.Body.String()),
		strings.TrimSpace(string(ceJSON))))
}

func Test_create_order_create_return_an_error(t *testing.T) {
	tests.MaybeFail = tests.InitFailFunc(t)

	ce := e.NewCustomHTTPStatus(e.StatusBadRequest)
	// ceJSON should be match by the response payload
	ceJSON, err := json.Marshal(ce)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// we do not care the input, just set OrderCreateService to return ce
	mockOrderService.EXPECT().OrderCreateService(gomock.Any()).Return(ce)

	st.GetRouter().HandleFunc("/v1/createorder", orderRestT.CreateOrder)
	request, _ := http.NewRequest(http.MethodPost, "/v1/createorder", nil)

	// Act
	recorder := httptest.NewRecorder()
	st.GetRouter().ServeHTTP(recorder, request)

	// Assert
	tests.MaybeFail("http_status", tests.Expect(recorder.Code, http.StatusOK))
	tests.MaybeFail("http_content_type", tests.Expect(
		recorder.Header().Get("Content-Type"), "application/json"))
	tests.MaybeFail("response_body", tests.Expect(
		strings.TrimSpace(recorder.Body.String()),
		strings.TrimSpace(string(ceJSON))))
}
