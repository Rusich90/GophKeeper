-- Удаление полей для синхронизации данных
ALTER TABLE users 
DROP COLUMN IF EXISTS encrypted_data,
DROP COLUMN IF EXISTS updated_at;