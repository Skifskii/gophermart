-- Создание таблицы для хранения информации о выводах средств
CREATE TABLE withdrawals (
    id SERIAL PRIMARY KEY,
    order_num VARCHAR(20) NOT NULL REFERENCES orders(number),
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_login VARCHAR(50) NOT NULL REFERENCES users(login)
);
