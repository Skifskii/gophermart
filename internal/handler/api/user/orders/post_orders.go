package orders

import (
	"errors"
	"gophermart/internal/middleware/authmw"
	om "gophermart/internal/service/orders"
	"io"
	"net/http"
)

type OrderAdder interface {
	AddOrder(userLogin, orderNum string) error
}

func NewPost(oa OrderAdder) http.HandlerFunc {
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

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
			return
		}

		orderNum := string(body)
		r.Body.Close()

		// Загружаем номер заказа в сервис
		if err := oa.AddOrder(userLogin, orderNum); err != nil {
			if errors.Is(err, om.ErrUserUploadedThisOrder) {
				// 200 - номер заказа уже был загружен этим пользователем
				w.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, om.ErrWrongOrderNumFormat) {
				// 422 — неверный формат номера заказа
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			if errors.Is(err, om.ErrAnotherUserUploadedThisOrder) {
				// 409 - номер заказа уже был загружен другим пользователем
				w.WriteHeader(http.StatusConflict)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 202 - новый номер заказа принят в обработку
		w.WriteHeader(http.StatusAccepted)
	}
}
