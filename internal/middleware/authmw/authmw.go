package authmw

import (
	"context"
	"net/http"
)

type ContextKey string

const UserLoginKey ContextKey = "user_login"

type UserLoginGetter interface {
	GetUserLogin(tokenString string) (login string, err error)
}

func AuthMiddleware(ug UserLoginGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, _ := r.Cookie("jwt")
			if cookie == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token := cookie.Value

			userLogin, err := ug.GetUserLogin(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserLoginKey, userLogin)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
