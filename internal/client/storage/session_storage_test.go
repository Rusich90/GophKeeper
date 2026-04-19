package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSessionStorage(t *testing.T) {
	ss := NewSessionStorage()
	if ss == nil {
		t.Fatal("NewSessionStorage() вернул nil")
	}

	if ss.sessionFile == "" {
		t.Error("NewSessionStorage() sessionFile пустой")
	}
}

func TestSessionStorage_SaveAndLoad(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем тестовую сессию
	session := &Session{
		Token:      "test-token-12345",
		Key:        "encryption-key-67890",
		LastSyncTS: 1234567890,
	}

	// Сохраняем
	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Проверяем, что файл создан
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("SaveSession() не создал файл")
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Проверяем данные
	if loaded.Token != "test-token-12345" {
		t.Errorf("LoadSession() Token = %v, хотим 'test-token-12345'", loaded.Token)
	}

	if loaded.Key != "encryption-key-67890" {
		t.Errorf("LoadSession() Key = %v, хотим 'encryption-key-67890'", loaded.Key)
	}

	if loaded.LastSyncTS != 1234567890 {
		t.Errorf("LoadSession() LastSyncTS = %v, хотим 1234567890", loaded.LastSyncTS)
	}
}

func TestSessionStorage_Load_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nonexistent.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Пытаемся загрузить несуществующий файл
	_, err := ss.LoadSession()
	if err == nil {
		t.Error("LoadSession() несуществующего файла должен возвращать ошибку")
	}

	expectedError := "session file not found"
	if err.Error() != expectedError {
		t.Errorf("LoadSession() ошибка = %v, хотим %v", err.Error(), expectedError)
	}
}

func TestSessionStorage_SaveAndLoad_MultipleSessions(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Сохраняем первую сессию
	session1 := &Session{
		Token:      "token-1",
		Key:        "key-1",
		LastSyncTS: 1000000000,
	}

	err := ss.SaveSession(session1)
	if err != nil {
		t.Fatalf("первый SaveSession() ошибка = %v", err)
	}

	// Перезаписываем второй сессией
	session2 := &Session{
		Token:      "token-2",
		Key:        "key-2",
		LastSyncTS: 2000000000,
	}

	err = ss.SaveSession(session2)
	if err != nil {
		t.Fatalf("второй SaveSession() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Должна быть вторая сессия
	if loaded.Token != "token-2" {
		t.Errorf("LoadSession() Token = %v, хотим 'token-2'", loaded.Token)
	}

	if loaded.Key != "key-2" {
		t.Errorf("LoadSession() Key = %v, хотим 'key-2'", loaded.Key)
	}

	if loaded.LastSyncTS != 2000000000 {
		t.Errorf("LoadSession() LastSyncTS = %v, хотим 2000000000", loaded.LastSyncTS)
	}
}

func TestSessionStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем и сохраняем сессию
	session := &Session{
		Token:      "test-token",
		Key:        "test-key",
		LastSyncTS: 1234567890,
	}

	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("файл не был создан")
	}

	// Удаляем
	err = ss.DeleteSession()
	if err != nil {
		t.Fatalf("DeleteSession() ошибка = %v", err)
	}

	// Проверяем, что файл удален
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("DeleteSession() не удалил файл")
	}
}

func TestSessionStorage_Delete_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nonexistent.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Удаляем несуществующий файл - не должно быть ошибки
	err := ss.DeleteSession()
	if err != nil {
		t.Errorf("DeleteSession() несуществующего файла вернул ошибку: %v", err)
	}
}

func TestSessionStorage_SaveAndLoad_EmptyFields(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем сессию с пустыми полями
	session := &Session{
		Token:      "",
		Key:        "",
		LastSyncTS: 0,
	}

	// Сохраняем
	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Проверяем, что поля пустые
	if loaded.Token != "" {
		t.Errorf("LoadSession() Token = %v, хотим пустую строку", loaded.Token)
	}

	if loaded.Key != "" {
		t.Errorf("LoadSession() Key = %v, хотим пустую строку", loaded.Key)
	}

	if loaded.LastSyncTS != 0 {
		t.Errorf("LoadSession() LastSyncTS = %v, хотим 0", loaded.LastSyncTS)
	}
}

func TestSessionStorage_SaveAndLoad_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем сессию со спецсимволами
	session := &Session{
		Token:      "token-with-special-chars-!@#$%%^&*()",
		Key:        "key-with-спецсимволы-привет-мир",
		LastSyncTS: 1234567890,
	}

	// Сохраняем
	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Проверяем данные
	if loaded.Token != "token-with-special-chars-!@#$%%^&*()" {
		t.Errorf("LoadSession() Token = %v, хотим 'token-with-special-chars-!@#$%%%%^&*()'", loaded.Token)
	}

	if loaded.Key != "key-with-спецсимволы-привет-мир" {
		t.Errorf("LoadSession() Key = %v, хотим 'key-with-спецсимволы-привет-мир'", loaded.Key)
	}
}

func TestSessionStorage_Load_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем поврежденный файл
	err := os.WriteFile(testFile, []byte("corrupted json data"), 0600)
	if err != nil {
		t.Fatalf("не удалось создать поврежденный файл: %v", err)
	}

	// Пытаемся загрузить
	_, err = ss.LoadSession()
	if err == nil {
		t.Error("LoadSession() поврежденного файла должен возвращать ошибку")
	}
}

func TestSessionStorage_SaveAndLoad_LongToken(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем сессию с очень длинным токеном
	longToken := ""
	for i := 0; i < 1000; i++ {
		longToken += "a"
	}

	session := &Session{
		Token:      longToken,
		Key:        "test-key",
		LastSyncTS: 1234567890,
	}

	// Сохраняем
	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Проверяем, что токен загружен полностью
	if len(loaded.Token) != 1000 {
		t.Errorf("LoadSession() длина Token = %v, хотим 1000", len(loaded.Token))
	}
}

func TestSessionStorage_SaveAndLoad_NegativeTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем сессию с отрицательным timestamp (теоретически возможно)
	session := &Session{
		Token:      "test-token",
		Key:        "test-key",
		LastSyncTS: -1234567890,
	}

	// Сохраняем
	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Проверяем timestamp
	if loaded.LastSyncTS != -1234567890 {
		t.Errorf("LoadSession() LastSyncTS = %v, хотим -1234567890", loaded.LastSyncTS)
	}
}

func TestSessionStorage_SaveAndLoad_MaxTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_session.json")

	ss := &SessionStorage{sessionFile: testFile}

	// Создаем сессию с максимальным timestamp
	session := &Session{
		Token:      "test-token",
		Key:        "test-key",
		LastSyncTS: 9223372036854775807, // Max int64
	}

	// Сохраняем
	err := ss.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession() ошибка = %v", err)
	}

	// Загружаем
	loaded, err := ss.LoadSession()
	if err != nil {
		t.Fatalf("LoadSession() ошибка = %v", err)
	}

	// Проверяем timestamp
	if loaded.LastSyncTS != 9223372036854775807 {
		t.Errorf("LoadSession() LastSyncTS = %v, хотим 9223372036854775807", loaded.LastSyncTS)
	}
}