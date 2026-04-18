package cli

import (
	"github.com/spf13/cobra"
)

// statusCmd представляет команду проверки статуса
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Проверка статуса авторизации",
	Long:  `Проверяет текущий статус авторизации пользователя.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем контейнер из контекста
		container, err := GetContainer(cmd)
		if err != nil {
			return err
		}

		// Проверяем наличие сессии (опционально)
		session, err := LoadSessionOptional(container)
		if err != nil {
			return err
		}

		if session == nil {
			container.UI.Output.Warning("Статус: Не авторизован")
			container.UI.Output.Plain("Для входа используйте: keeper login")
			return nil
		}

		container.UI.Output.Success("Статус: Авторизован")
		if session.Key != "" {
			container.UI.Output.Plain("Ключ шифрования: сохранен")
		} else {
			container.UI.Output.Warning("Ключ шифрования: не сохранен")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
