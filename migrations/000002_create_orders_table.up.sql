-- Создание таблицы заказов
CREATE TABLE orders (
    number VARCHAR(20) PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    accrual NUMERIC(10, 2),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_login VARCHAR(50) NOT NULL REFERENCES users(login)
);