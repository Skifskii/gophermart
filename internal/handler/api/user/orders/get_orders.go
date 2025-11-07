package orders

import (
	"encoding/json"
	"gophermart/internal/middleware/authmw"
	"gophermart/internal/model"
	"net/http"
	"time"
)

type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

type OrdersGetter interface {
	GetOrders(userLogin string) ([]model.Order, error)
}

func NewGet(og OrdersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Получаем логин пользователя
		userLogin, ok := r.Context().Value(authmw.UserLoginKey).(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Получаем список заказов
		orders, err := og.GetOrders(userLogin)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		resp := make([]OrderResponse, len(orders))
		for i, o := range orders {
			resp[i] = toOrderResponse(o)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(resp)

		w.WriteHeader(http.StatusOK)
	}
}

func toOrderResponse(o model.Order) OrderResponse {
	return OrderResponse{
		Number:     o.Number,
		Status:     string(o.Status),
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}
}
