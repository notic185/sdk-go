package sdk

import (
	"net/http"
	"testing"
	"time"
)

var m, _ = New(
	"http://10.0.0.254:4003",
	UserCredential{
		ExternalId: "sbSO9eCvjQkE38hVnrVy4TiD",
		Secret:     "ER3wT3UP05TVEyh8CdMnZsCz5I1j0z",
	},
)

func TestMangrove_CreateMerchantOrder(t *testing.T) {
	// -
	merchantOrder, wrong := m.MerchantOrder.Create([]MerchantOrder{
		{
			Order: Order{
				// 1 分钱
				Amount: 1,
				OrderCallback: OrderCallback{
					Endpoint: "http://127.0.0.1:8888",
				},
			},
		},
	})
	// -
	if wrong == nil {
		t.Logf("The code for this order is %s", merchantOrder[0].Order.Code)
	} else {
		t.Error(wrong)
	}
}

func TestMangrove_DescribeOrder(t *testing.T) {
	// -
	order, wrong := m.Order.Describe("05bf25dd-103a-4580-96b2-4e30c9736822")
	// -
	if wrong == nil {
		t.Logf("The code for this order is %s", order.Code)
	} else {
		t.Error(wrong)
	}
}

func TestMangrove_UpdateOrder(t *testing.T) {
	// -
	orders, wrong := m.Order.Update(Order{
		Model: Model{
			UUID: "05bf25dd-103a-4580-96b2-4e30c9736822",
		},
		NamedModel: NamedModel{
			Name: time.Now().String(),
		},
	})
	// -
	if wrong == nil {
		t.Logf("The code for this order is %s", orders[0].Code)
	} else {
		t.Error(wrong)
	}
}

func TestMangrove_DeleteOrder(t *testing.T) {
	if wrong := m.Order.Delete("15bf25dd-103a-4580-96b2-4e30c9736822"); wrong != nil {
		t.Error(wrong)
	}
}

func TestMangrove_HandleOrderCallback(t *testing.T) {
	// -
	wrong := http.ListenAndServe(":8888", http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// -
		requestPayload, wrong := m.Order.HandleCallback(request)
		// -
		if wrong == nil {
			t.Logf("The code for this order is %s", requestPayload.Code)
		} else {
			t.Error(wrong)
		}
	}))
	// -
	if wrong != nil {
		t.Error(wrong)
	}
}

func TestMangrove_SummarizeUserIntegralAmount(t *testing.T) {
	// -
	result, wrong := m.User.SummarizeIntegralAmount()
	// -
	if wrong == nil {
		for key, value := range result {
			t.Logf("%s → %f", key, value)
		}
	} else {
		t.Error(wrong)
	}
}

func TestMangrove_CreateUserOrder(t *testing.T) {
	// -
	userOrders, wrong := m.UserOrder.Create([]UserOrder{
		{
			Currency: "1",
			Order: Order{
				OrderCallback: OrderCallback{
					Endpoint: "http://127.0.0.1:8888",
				},
			},
			UserOrderTransaction: UserOrderTransaction{
				Amount: 1,
			},
		},
	})
	// -
	if wrong == nil {
		t.Logf("The code for this order is %s", userOrders[0].Order.Code)
	} else {
		t.Error(wrong)
	}
}
