package register

import (
	"encoding/json"
	"errors"
	"gophermart/internal/model"
	"gophermart/internal/repository"
	"net/http"
)

type UserCreator interface {
	CreateUser(login, password string) (jwtToken string, err error)
}

func New(uc UserCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// десериализуем запрос в структуру
		var req model.UserAuth
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Проверяем формат запроса
		if ok := validateReq(req); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Регистрируем пользователя
		jwtToken, err := uc.CreateUser(req.Login, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrLoginAlreadyTaken) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "jwt",
			Value: jwtToken,
		})

		w.WriteHeader(http.StatusOK)
	}
}

func validateReq(req model.UserAuth) bool {
	if req.Login == "" || req.Password == "" {
		return false
	}

	return true
}
