package withdraw

import (
	"encoding/json"
	"errors"
	"gophermart/internal/middleware/authmw"
	"gophermart/internal/model"
	"gophermart/internal/repository"
	"gophermart/internal/service/orders"
	"gophermart/internal/service/withdrawal"
	"net/http"
)

type balanceManager interface {
	GetBalance(login string) (current, withdrawn float64, err error)
	UpdateBalance(login string, amount float64) error
}

type orderManager interface {
	GetOrder(orderNum string) (model.Order, error)
}

type WithdrawalRecorder interface {
	RecordWithdrawal(userLogin, orderNum string, amount float64) error
}

func NewPost(wr WithdrawalRecorder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Получаем логин пользователя
		userLogin, ok := r.Context().Value(authmw.UserLoginKey).(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// десериализуем запрос в структуру
		var req WithdrawRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Загружаем номер заказа в сервис
		if err := wr.RecordWithdrawal(userLogin, req.Order, -1*req.Sum); err != nil { // TODO: проверить работу.
			if errors.Is(err, withdrawal.ErrInsufficientFunds) {
				// 402 — на счету недостаточно средств
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			if errors.Is(err, repository.ErrOrderNotFound) || errors.Is(err, orders.ErrAnotherUserUploadedThisOrder) {
				// 422 — неверный номер заказа;
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 200 — успешная обработка запроса
		w.WriteHeader(http.StatusOK)
	}
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
