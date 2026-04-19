package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd представляет команду удаления записи
var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Удалить запись",
	Long: `Удалить запись из хранилища секретов.
	
Команда находит запись по началу ID и удаляет её.
Аргумент id - это начало идентификатора записи (минимум 1 символ).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем контейнер из контекста
		container, err := GetContainer(cmd)
		if err != nil {
			return err
		}

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

		// Поиск записи по префиксу ID через сервис
		idPrefix := args[0]
		item, duplicates, err := container.SecretService.FindSecretByIDPrefix(cmd.Context(), storageData, idPrefix)

		// Обработка результатов поиска
		if err != nil {
			container.UI.Output.Errorf("%v", err)
			return err
		}

		if len(duplicates) > 1 {
			// Найдено несколько записей - выводим таблицу дубликатов
			container.UI.Output.Plain(fmt.Sprintf("Найдено несколько записей с префиксом '%s', уточните запрос:", idPrefix))
			container.UI.TableRenderer.RenderDuplicatesTable(duplicates)
			return nil
		}

		// Удаляем секрет через сервис
		if err := container.SecretService.DeleteSecret(cmd.Context(), storageData, item.ID); err != nil {
			container.UI.Output.Errorf("Ошибка удаления секрета: %v", err)
			return fmt.Errorf("ошибка удаления секрета: %w", err)
		}

		// Сохраняем секреты
		if err := container.SecretStorage.Save(storageData, session.Key); err != nil {
			container.UI.Output.Errorf("Ошибка сохранения секретов: %v", err)
			return fmt.Errorf("ошибка сохранения секретов: %w", err)
		}

		container.UI.Output.Success(fmt.Sprintf("Запись '%s' успешно удалена", item.Title))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
