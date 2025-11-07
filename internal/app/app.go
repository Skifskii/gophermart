package app

import (
	"fmt"
	"gophermart/internal/config"
	"gophermart/internal/handler/api/user/balance"
	"gophermart/internal/handler/api/user/balance/withdraw"
	"gophermart/internal/handler/api/user/login"
	"gophermart/internal/handler/api/user/orders"

	"gophermart/internal/handler/api/user/register"
	"gophermart/internal/middleware/authmw"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service/auth"
	bm "gophermart/internal/service/balance"
	om "gophermart/internal/service/orders"
	"gophermart/internal/service/orders/loyalsys"
	"gophermart/internal/service/withdrawal"
	"net/http"

	"github.com/go-chi/chi"
)

func Run() error {
	// Конфиг
	cfg := config.New()

	// Логгер
	// TODO:

	// Репозиторий
	repo, _ := postgres.New(cfg.DatabaseURI)

	// Сервисы:
	// - сервис авторизации
	authService := auth.New(repo, cfg.SecretKey)

	// - сервис управления балансом
	balanceManager := bm.New(repo)

	// - сервис управления заказами
	loyaltySystem := loyalsys.New()
	ordersManager := om.New(repo, loyaltySystem)
	withdrawalManager := withdrawal.New(repo, balanceManager, ordersManager)

	// HTTP сервер
	router := chi.NewRouter()

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", register.New(authService))
		r.Post("/login", login.New(authService))

		r.Group(func(r chi.Router) {
			r.Use(authmw.AuthMiddleware(authService))

			r.Get("/balance", balance.New(balanceManager))
			r.Post("/balance/withdraw", withdraw.NewPost(withdrawalManager))
			r.Post("/orders", orders.NewPost(ordersManager))
			r.Get("/orders", orders.NewGet(ordersManager))
		})
	})

	fmt.Printf("Starting server at %s\n", cfg.RunAddress)
	http.ListenAndServe(cfg.RunAddress, router)

	return nil
}
