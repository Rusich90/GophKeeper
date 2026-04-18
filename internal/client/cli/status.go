package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
	"github.com/spf13/cobra"
)

// statusCmd представляет команду проверки статуса
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Проверка статуса авторизации",
	Long:  `Проверяет текущий статус авторизации пользователя.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем UI из контекста
		uiInstance, ok := cmd.Context().Value("ui").(*ui.UI)
		if !ok {
			return fmt.Errorf("UI не инициализирован")
		}

		// Получаем хранилище сессий из контекста
		sessionStorage, ok := cmd.Context().Value("sessionStorage").(*storage.SessionStorage)
		if !ok {
			return fmt.Errorf("хранилище сессий не инициализировано")
		}

		// Проверяем наличие сессии
		session, err := sessionStorage.LoadSession()
		if err != nil {
			uiInstance.Output.Warning("Статус: Не авторизован")
			uiInstance.Output.Plain("Для входа используйте: keeper login")
			return nil
		}

		uiInstance.Output.Success("Статус: Авторизован")
		if session.Key != "" {
			uiInstance.Output.Plain("Ключ шифрования: сохранен")
		} else {
			uiInstance.Output.Warning("Ключ шифрования: не сохранен")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
