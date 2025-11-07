package withdrawals

import (
	"encoding/json"
	"gophermart/internal/middleware/authmw"
	"gophermart/internal/model"
	"net/http"
	"time"
)

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type WithdrawalsGetter interface {
	GetWithdrawals(userLogin string) ([]model.Withdrawal, error)
}

func NewGet(wg WithdrawalsGetter) http.HandlerFunc {
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
		withdrawals, err := wg.GetWithdrawals(userLogin)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		resp := make([]WithdrawalResponse, len(withdrawals))
		for i, o := range withdrawals {
			resp[i] = toWithdrawalResponse(o)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(resp)

		w.WriteHeader(http.StatusOK)
	}
}

func toWithdrawalResponse(w model.Withdrawal) WithdrawalResponse {
	return WithdrawalResponse{
		Order:       w.Order,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	}
}
