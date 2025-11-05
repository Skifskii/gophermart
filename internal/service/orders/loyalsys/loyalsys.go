package loyalsys

import (
	"errors"
	"gophermart/internal/model"
	"time"
)

var ErrOrderNotRegistered = errors.New("the order is not registered in the payment system")
var ErrRequestLimitReached = errors.New("the number of requests to the service has been exceeded") // TODO: может возвращаться при запросе в смежный сервис

type LoyaltySystem struct{}

func New() *LoyaltySystem {
	return &LoyaltySystem{}
}

type Status string

const (
	StatusRegistered Status = "REGISTERED"
	StatusInvalid    Status = "INVALID"
	StatusProcessing Status = "PROCESSING"
	StatusProcessed  Status = "PROCESSED"
)

func (s *Status) toDomain() model.Status {
	switch *s {
	case StatusProcessing:
		return model.StatusProcessing
	case StatusInvalid:
		return model.StatusInvalid
	case StatusProcessed:
		return model.StatusProcessed
	default:
		return model.StatusNew
	}
}

type orderResponse struct {
	Order   string   `json:"order"`
	Status  Status   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

func (or *orderResponse) toDomain() model.Order {
	return model.Order{
		Number:     or.Order,
		Status:     or.Status.toDomain(),
		Accrual:    or.Accrual,
		UploadedAt: time.Time{},
	}
}

func (ls *LoyaltySystem) GetOrderInfo(orderNum string) (model.Order, error) {
	f500 := 500.
	// TODO: прикрутить сервис
	mockOrders := []orderResponse{
		{
			Order:   "123",
			Status:  "PROCESSED",
			Accrual: &f500,
		},
		{
			Order:   "456",
			Status:  "PROCESSED",
			Accrual: &f500,
		},
		{
			Order:  "789",
			Status: "REGISTERED",
		},
	}

	for _, order := range mockOrders {
		if order.Order == orderNum {
			return order.toDomain(), nil
		}
	}

	return model.Order{}, ErrOrderNotRegistered
}
