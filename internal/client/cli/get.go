package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// getCmd представляет команду просмотра детальной информации записи
var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Просмотреть детальную информацию записи",
	Long: `Просмотреть детальную информацию конкретной записи.
	
Команда находит запись по началу ID и раскрывает секретные данные (пароли, CVV).
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

		// Выводим детальную информацию
		switch item.Type {
		case "login":
			printLoginDetails(container, *item)
		case "text":
			printTextDetails(container, *item)
		case "card":
			printCardDetails(container, *item)
		default:
			container.UI.Output.Errorf("Неизвестный тип записи: %s", item.Type)
			return fmt.Errorf("неизвестный тип записи: %s", item.Type)
		}

		return nil
	},
}

// printLoginDetails выводит детальную информацию записи типа login
func printLoginDetails(container *Container, item storage.Item) {
	container.UI.Output.Plainf("Title: %s", item.Title)
	container.UI.Output.Plainf("Type: %s", item.Type)
	container.UI.Output.Plainf("ID: %s", item.ID)
	container.UI.Output.Plainf("Username: %s", item.Username)
	container.UI.Output.Plainf("Password: %s", item.Password)
	container.UI.Output.Plainf("Updated: %s", item.UpdatedAt)
}

// printTextDetails выводит детальную информацию записи типа text
func printTextDetails(container *Container, item storage.Item) {
	container.UI.Output.Plainf("Title: %s", item.Title)
	container.UI.Output.Plainf("Type: %s", item.Type)
	container.UI.Output.Plainf("ID: %s", item.ID)
	container.UI.Output.Plain("Content:")
	container.UI.Output.Plain(item.Content)
	container.UI.Output.Plainf("Updated: %s", item.UpdatedAt)
}

// printCardDetails выводит детальную информацию записи типа card
func printCardDetails(container *Container, item storage.Item) {
	container.UI.Output.Plainf("Title: %s", item.Title)
	container.UI.Output.Plainf("Type: %s", item.Type)
	container.UI.Output.Plainf("ID: %s", item.ID)
	container.UI.Output.Plainf("Number: %s", item.Number)
	container.UI.Output.Plainf("Expiry: %s/%s", item.ExpiryMonth, item.ExpiryYear)
	container.UI.Output.Plainf("CVV: %s", item.CVV)
	container.UI.Output.Plainf("Updated: %s", item.UpdatedAt)
}

func init() {
	rootCmd.AddCommand(getCmd)
}
