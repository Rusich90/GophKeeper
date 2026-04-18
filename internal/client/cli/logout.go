package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// logoutCmd представляет команду выхода
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Выход из системы",
	Long:  `Выход из системы и удаление сохраненной сессии.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем контейнер из контекста
		container, err := GetContainer(cmd)
		if err != nil {
			return err
		}

		// Проверяем, есть ли сессия (опционально)
		session, err := LoadSessionOptional(container)
		if err != nil {
			return err
		}

		if session == nil {
			container.UI.Output.Warning("Вы не авторизованы")
			return nil
		}

		// Вызываем серверный метод logout
		if err := container.AuthService.Logout(cmd.Context(), session.Token); err != nil {
			container.UI.Output.Warning(fmt.Sprintf("Не удалось уведомить сервер о выходе: %v", err))
			// Продолжаем удаление локальной сессии даже если сервер недоступен
		}

		// Удаляем сессию локально
		if err := container.SessionStorage.DeleteSession(); err != nil {
			container.UI.Output.Errorf("Ошибка при выходе: %v", err)
			return fmt.Errorf("ошибка при выходе: %w", err)
		}

		container.UI.Output.Success("Выход выполнен успешно")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
