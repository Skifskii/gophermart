-- Создание таблицы пользователей
CREATE TABLE users (
    login VARCHAR(50) PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    balance NUMERIC(15, 2) DEFAULT 0.00,
    withdrawn NUMERIC(15, 2) DEFAULT 0.00
);