package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// statusCmd представляет команду проверки статуса
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Проверка статуса авторизации",
	Long:  `Проверяет текущий статус авторизации пользователя.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем output из контекста
		output, ok := cmd.Context().Value("output").(*ColorOutput)
		if !ok {
			return fmt.Errorf("сервис вывода не инициализирован")
		}

		// Инициализируем хранилище сессий
		sessionStorage := storage.NewSessionStorage()

		// Проверяем наличие сессии
		session, err := sessionStorage.LoadSession()
		if err != nil {
			output.Warning("Статус: Не авторизован")
			output.Plain("Для входа используйте: keeper login")
			return nil
		}

		output.Success("Статус: Авторизован")
		if session.Key != "" {
			output.Plain("Ключ шифрования: сохранен")
		} else {
			output.Warning("Ключ шифрования: не сохранен")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
