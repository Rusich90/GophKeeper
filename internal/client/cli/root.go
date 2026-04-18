package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Rusich90/GophKeeper/internal/client/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	serverAddr string
	verbose    bool
)

// rootCmd представляет базовую команду CLI приложения
var rootCmd = &cobra.Command{
	Use:   "keeper",
	Short: "GophKeeper CLI - клиент для управления секретами",
	Long: `GophKeeper - это безопасное хранилище секретов.
CLI клиент позволяет управлять вашими данными через командную строку.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Инициализация конфигурации
		var cfg *config.Config
		var err error

		if cfgFile != "" {
			// Используем указанный файл конфигурации
			cfg, err = config.InitConfigWithPath(cfgFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка загрузки конфигурации: %v\n", err)
				return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
			}
		} else {
			// Используем стандартный файл конфигурации в домашней директории
			cfg, err = config.InitConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка инициализации конфигурации: %v\n", err)
				return fmt.Errorf("ошибка инициализации конфигурации: %w", err)
			}
		}

		// Переопределяем адрес сервера из флага командной строки, если указан
		if serverAddr != "" {
			cfg.ServerAddr = serverAddr
		}

		// Создаем контейнер зависимостей
		container, err := NewContainer(cfg, verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка инициализации контейнера: %v\n", err)
			return fmt.Errorf("ошибка инициализации контейнера: %w", err)
		}

		// Сохраняем контейнер в контексте команды
		ctx := context.WithValue(cmd.Context(), containerKey{}, container)
		cmd.SetContext(ctx)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Закрытие соединений после выполнения команды
		if container, ok := cmd.Context().Value(containerKey{}).(*Container); ok {
			return container.Close()
		}
		return nil
	},
}

// Execute запускает CLI приложение
// Эта функция является точкой входа для CLI клиента GophKeeper.
// Она инициализирует все команды и обрабатывает выполнение.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Глобальные флаги
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "файл конфигурации (по умолчанию $HOME/.keeper.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "", "адрес сервера (переопределяет значение из конфигурации)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "подробный вывод")
}
