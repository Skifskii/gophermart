package loyalsys

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/internal/model"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrOrderNotRegistered = errors.New("the order is not registered in the payment system")
var ErrRequestLimitReached = errors.New("the number of requests to the service has been exceeded") // TODO: может возвращаться при запросе в смежный сервис

type LoyaltySystem struct {
	address string
}

func New(address string) *LoyaltySystem {
	return &LoyaltySystem{address: address}
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
	url := strings.TrimRight(ls.address, "/") + "/api/orders/" + orderNum

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return model.Order{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return model.Order{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var or orderResponse
		if err := json.NewDecoder(resp.Body).Decode(&or); err != nil {
			return model.Order{}, err
		}
		return or.toDomain(), nil
	case http.StatusNotFound:
		return model.Order{}, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		return model.Order{}, ErrRequestLimitReached
	default:
		body, _ := io.ReadAll(resp.Body)
		return model.Order{}, fmt.Errorf("loyalsys: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func (ls *LoyaltySystem) mockGetOrderInfo(orderNum string) (model.Order, error) {
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
