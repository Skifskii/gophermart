package postgres

import (
	"database/sql"
	"errors"
	"gophermart/internal/model"
	"gophermart/internal/repository"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

var errEmptyDSN = errors.New("DSN is empty")
var errEmptyPassword = errors.New("password is empty")

type Repo struct {
	db *sql.DB
}

func New(dsn string) (*Repo, error) {
	if dsn == "" {
		return nil, errEmptyDSN
	}

	// Запускаем миграции
	if err := runMigration(dsn); err != nil {
		return nil, err
	}

	// Подключаемся к БД
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	repo := &Repo{
		db: db,
	}

	return repo, nil
}

func runMigration(dsn string) error {
	m, err := migrate.New(
		"file://./migrations",
		dsn,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (r *Repo) AddUser(login, password string) error {
	pwdHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		login,
		pwdHash,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrLoginAlreadyTaken
		}
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	if password == "" {
		return "", errEmptyPassword
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func (r *Repo) AuthenticateUser(login, password string) error {
	var hashedPwdFromDB string
	err := r.db.QueryRow(
		"SELECT password_hash FROM users WHERE login = $1",
		login,
	).Scan(&hashedPwdFromDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrUserLoginNotFound
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPwdFromDB), []byte(password))
	if err != nil {
		return repository.ErrWrongPassword
	}

	return nil
}

func (r *Repo) GetBalance(login string) (current, withdrawn float64, err error) {
	err = r.db.QueryRow(
		"SELECT balance, withdrawn FROM users WHERE login = $1",
		login,
	).Scan(&current, &withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, repository.ErrUserLoginNotFound
		}
		return 0, 0, err
	}

	return current, withdrawn, err
}

func (r *Repo) GetOrders(login string) ([]model.Order, error) {
	// Получаем заказы пользователя
	rows, err := r.db.Query(
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE user_login = $1;",
		login,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserLoginNotFound
		}
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var ord model.Order
		err := rows.Scan(&ord.Number, &ord.Status, &ord.Accrual, &ord.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, ord)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repo) GetUnfinishedOrders() ([]model.Order, error) {
	// Получаем заказы пользователя
	rows, err := r.db.Query(
		"SELECT number, status, accrual, uploaded_at, user_login FROM orders WHERE status = 'NEW' OR status = 'PROCESSING';",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var ord model.Order
		err := rows.Scan(&ord.Number, &ord.Status, &ord.Accrual, &ord.UploadedAt, &ord.UserLogin)
		if err != nil {
			return nil, err
		}
		orders = append(orders, ord)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repo) AddOrder(login string, order model.Order) error {
	_, err := r.db.Exec(
		"INSERT INTO orders (number, status, accrual, uploaded_at, user_login) VALUES ($1, $2, $3, $4, $5)",
		order.Number, order.Status, order.Accrual, order.UploadedAt,
		login,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrOrderAlreadyExists
		}
	}

	return err
}

func (r *Repo) UpdateOrders(orders []model.Order) error {
	if len(orders) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updOrderStmt, err := tx.Prepare(
		`UPDATE orders
		 SET status = $1, accrual = $2
		 WHERE number = $3`,
	)
	if err != nil {
		return err
	}
	defer updOrderStmt.Close()

	updUserStmt, err := tx.Prepare(
		`UPDATE users
		 SET balance = balance + $1
		 WHERE login = $2`,
	)
	if err != nil {
		return err
	}
	defer updOrderStmt.Close()

	for _, ord := range orders {
		_, err := updOrderStmt.Exec(ord.Status, ord.Accrual, ord.Number)
		if err != nil {
			return err
		}

		if ord.Accrual != nil && *ord.Accrual > 0 {
			_, err = updUserStmt.Exec(*ord.Accrual, ord.UserLogin)
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetOrder(orderNum string) (ord model.Order, err error) {
	err = r.db.QueryRow(
		"SELECT number, status, accrual, uploaded_at, user_login FROM orders WHERE number = $1",
		orderNum,
	).Scan(&ord.Number, &ord.Status, &ord.Accrual, &ord.UploadedAt, &ord.UserLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ord, repository.ErrOrderNotFound
		}
	}

	return ord, err
}
