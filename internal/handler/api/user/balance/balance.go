package balance

import (
	"encoding/json"
	"gophermart/internal/middleware/authmw"
	"gophermart/internal/model"
	"net/http"
)

type BalanceGetter interface {
	GetBalance(login string) (current, withdrawn float64, err error)
}

func New(bg BalanceGetter) http.HandlerFunc {
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

		// Получаем баланс по логину
		current, withdrawn, err := bg.GetBalance(userLogin)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		resp := model.BalanceResponse{
			Current:   current,
			Withdrawn: withdrawn,
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(resp)

		w.WriteHeader(http.StatusOK)
	}
}
