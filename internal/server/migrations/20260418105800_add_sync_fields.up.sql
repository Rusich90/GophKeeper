-- Добавление полей для синхронизации данных
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS encrypted_data BYTEA,
ADD COLUMN IF NOT EXISTS updated_at BIGINT DEFAULT 0;