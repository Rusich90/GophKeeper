package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config представляет конфигурацию клиента
type Config struct {
	ServerAddr string `yaml:"server_addr"`
}

// InitConfig инициализирует конфигурацию из файла конфигурации
// Если файл не существует, создает его с дефолтными значениями
func InitConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения домашней директории: %w", err)
	}

	configPath := filepath.Join(homeDir, ".keeper.yaml")

	// Проверяем существование файла конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Создаем файл с дефолтными значениями
		defaultConfig := &Config{
			ServerAddr: "127.0.0.1:50051",
		}

		if err := saveConfig(configPath, defaultConfig); err != nil {
			return nil, fmt.Errorf("ошибка создания файла конфигурации: %w", err)
		}

		return defaultConfig, nil
	}

	// Читаем существующий файл конфигурации
	return loadConfig(configPath)
}

// InitConfigWithPath инициализирует конфигурацию из указанного файла
func InitConfigWithPath(configPath string) (*Config, error) {
	// Проверяем существование файла конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл конфигурации не найден: %s", configPath)
	}

	return loadConfig(configPath)
}

// loadConfig загружает конфигурацию из файла
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга файла конфигурации: %w", err)
	}

	// Устанавливаем значение по умолчанию, если поле пустое
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = "127.0.0.1:50051"
	}

	return &cfg, nil
}

// saveConfig сохраняет конфигурацию в файл
func saveConfig(configPath string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("ошибка сериализации конфигурации: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("ошибка записи файла конфигурации: %w", err)
	}

	return nil
}

// UpdateServerAddr обновляет адрес сервера в конфигурации
func UpdateServerAddr(configPath string, serverAddr string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	cfg.ServerAddr = serverAddr
	return saveConfig(configPath, cfg)
}
