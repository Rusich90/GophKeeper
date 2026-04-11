-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users
(
  login         VARCHAR(50) PRIMARY KEY,
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
