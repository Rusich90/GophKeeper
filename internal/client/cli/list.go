package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd представляет команду вывода списка записей
var listCmd = &cobra.Command{
	Use:   "list [type]",
	Short: "Вывести список сохраненных данных",
	Long: `Вывести список сохраненных данных указанного типа.
	
Поддерживаемые типы записей:
  - login: логин и пароль
  - text: текстовая заметка
  - card: данные банковской карты`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем контейнер из контекста
		container, err := GetContainer(cmd)
		if err != nil {
			return err
		}

		// Получаем тип данных
		dataType := args[0]

		// Загружаем сессию
		session, err := LoadSession(container)
		if err != nil {
			return err
		}

		// Загружаем секреты
		storageData, err := container.SecretStorage.Load(session.Key)
		if err != nil {
			container.UI.Output.Errorf("Ошибка загрузки секретов: %v", err)
			return fmt.Errorf("ошибка загрузки секретов: %w", err)
		}

		// Фильтруем секреты по типу через сервис
		filteredItems, err := container.SecretService.FilterSecretsByType(cmd.Context(), storageData, dataType)
		if err != nil {
			container.UI.Output.Errorf("%v", err)
			return err
		}

		// Проверяем, есть ли секреты
		if len(filteredItems) == 0 {
			container.UI.Output.Plain("Секреты не найдены")
			return nil
		}

		// Диспетчеризация функций рендеринга
		switch dataType {
		case "login":
			container.UI.TableRenderer.RenderLoginTable(filteredItems)
		case "text":
			container.UI.TableRenderer.RenderTextTable(filteredItems)
		case "card":
			container.UI.TableRenderer.RenderCardTable(filteredItems)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
