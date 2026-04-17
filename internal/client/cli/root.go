package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Rusich90/GophKeeper/internal/client/config"
	"github.com/Rusich90/GophKeeper/internal/client/grpc"
	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	serverAddr string
	verbose    bool
	output     *ColorOutput
)

// rootCmd представляет базовую команду CLI приложения
var rootCmd = &cobra.Command{
	Use:   "keeper",
	Short: "GophKeeper CLI - клиент для управления секретами",
	Long: `GophKeeper - это безопасное хранилище секретов.
CLI клиент позволяет управлять вашими данными через командную строку.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Инициализация вывода
		output = NewColorOutput(verbose)

		// Инициализация конфигурации
		var cfg *config.Config
		var err error

		if cfgFile != "" {
			// Используем указанный файл конфигурации
			cfg, err = config.InitConfigWithPath(cfgFile)
			if err != nil {
				output.Errorf("Ошибка загрузки конфигурации: %v", err)
				return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
			}
		} else {
			// Используем стандартный файл конфигурации в домашней директории
			cfg, err = config.InitConfig()
			if err != nil {
				output.Errorf("Ошибка инициализации конфигурации: %v", err)
				return fmt.Errorf("ошибка инициализации конфигурации: %w", err)
			}
		}

		// Переопределяем адрес сервера из флага командной строки, если указан
		if serverAddr != "" {
			cfg.ServerAddr = serverAddr
		}

		// Инициализация gRPC клиента
		grpcClient, err := grpc.NewClient(cfg.ServerAddr)
		if err != nil {
			output.Errorf("Ошибка подключения к серверу: %v", err)
			return fmt.Errorf("ошибка подключения к серверу: %w", err)
		}

		// Сохраняем клиента в контексте команды
		ctx := context.WithValue(cmd.Context(), "grpcClient", grpcClient)
		ctx = context.WithValue(ctx, "config", cfg)
		ctx = context.WithValue(ctx, "output", output)

		// Инициализация сервиса авторизации
		authService := service.NewAuthService(grpcClient)
		ctx = context.WithValue(ctx, "authService", authService)

		// Инициализация хранилища сессий
		sessionStorage := storage.NewSessionStorage()
		ctx = context.WithValue(ctx, "sessionStorage", sessionStorage)

		cmd.SetContext(ctx)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Закрытие соединений после выполнения команды
		if authService, ok := cmd.Context().Value("authService").(service.AuthService); ok {
			return authService.Close()
		}
		return nil
	},
}

// Execute запускает CLI приложение
// Эта функция является точкой входа для CLI клиента GophKeeper.
// Она инициализирует все команды и обрабатывает выполнение.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if output == nil {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		} else {
			output.Errorf("%v", err)
		}
		os.Exit(1)
	}
}

func init() {
	// Глобальные флаги
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "файл конфигурации (по умолчанию $HOME/.keeper.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "", "адрес сервера (переопределяет значение из конфигурации)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "подробный вывод")
}
