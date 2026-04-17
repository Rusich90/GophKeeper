package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// logoutCmd представляет команду выхода
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Выход из системы",
	Long:  `Выход из системы и удаление сохраненной сессии.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем output из контекста
		output, ok := cmd.Context().Value("output").(*ColorOutput)
		if !ok {
			return fmt.Errorf("сервис вывода не инициализирован")
		}

		// Получаем сервис авторизации из контекста
		authService, ok := cmd.Context().Value("authService").(service.AuthService)
		if !ok {
			return fmt.Errorf("сервис авторизации не инициализирован")
		}

		// Инициализируем хранилище сессий
		sessionStorage := storage.NewSessionStorage()

		// Проверяем, есть ли сессия
		session, err := sessionStorage.LoadSession()
		if err != nil {
			output.Warning("Вы не авторизованы")
			return nil
		}

		// Вызываем серверный метод logout
		if err := authService.Logout(cmd.Context(), session.Token); err != nil {
			output.Warning(fmt.Sprintf("Не удалось уведомить сервер о выходе: %v", err))
			// Продолжаем удаление локальной сессии даже если сервер недоступен
		}

		// Удаляем сессию локально
		if err := sessionStorage.DeleteSession(); err != nil {
			output.Errorf("Ошибка при выходе: %v", err)
			return fmt.Errorf("ошибка при выходе: %w", err)
		}

		output.Success("Выход выполнен успешно")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
