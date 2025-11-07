package login

import (
	"encoding/json"
	"gophermart/internal/model"
	"net/http"
)

type UserAuthenticator interface {
	AuthenticateUser(login, password string) (jwtToken string, err error)
}

func New(ua UserAuthenticator) http.HandlerFunc {
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

		// Проверяем пользователя
		jwtToken, err := ua.AuthenticateUser(req.Login, req.Password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
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
