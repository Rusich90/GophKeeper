package cli

import (
	"fmt"
	"strings"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
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

		// Получаем ключ шифрования из сессии
		session, err := sessionStorage.LoadSession()
		if err != nil {
			uiInstance.Output.Error("Сессия не инициализирована. Пожалуйста, войдите в систему.")
			return fmt.Errorf("сессия не инициализирована: %w", err)
		}

		// Загружаем данные
		dataStorage := storage.NewDataStorage()
		storageData, err := dataStorage.Load(session.Key)
		if err != nil {
			uiInstance.Output.Errorf("Ошибка загрузки данных: %v", err)
			return fmt.Errorf("ошибка загрузки данных: %w", err)
		}

		// Поиск записи по префиксу ID
		idPrefix := args[0]
		var foundItems []storage.Item
		
		for _, item := range storageData.Items {
			if strings.HasPrefix(item.ID, idPrefix) {
				foundItems = append(foundItems, item)
			}
		}

		// Обработка результатов поиска
		switch len(foundItems) {
		case 0:
			uiInstance.Output.Errorf("Запись с ID '%s' не найдена", idPrefix)
			return fmt.Errorf("запись не найдена")
			
		case 1:
			// Выводим детальную информацию
			item := foundItems[0]
			switch item.Type {
			case "login":
				printLoginDetails(item)
			case "text":
				printTextDetails(item)
			case "card":
				printCardDetails(item)
			default:
				uiInstance.Output.Errorf("Неизвестный тип записи: %s", item.Type)
				return fmt.Errorf("неизвестный тип записи: %s", item.Type)
			}
			
		default:
			// Найдено несколько записей - выводим таблицу дубликатов
			uiInstance.Output.Plain(fmt.Sprintf("Найдено несколько записей с префиксом '%s', уточните запрос:", idPrefix))
			uiInstance.TableRenderer.RenderDuplicatesTable(foundItems)
		}

		return nil
	},
}

// printLoginDetails выводит детальную информацию записи типа login
func printLoginDetails(item storage.Item) {
	fmt.Printf("Title: %s\n", item.Title)
	fmt.Printf("Type: %s\n", item.Type)
	fmt.Printf("ID: %s\n", item.ID)
	fmt.Printf("Username: %s\n", item.Username)
	fmt.Printf("Password: %s\n", item.Password)
	fmt.Printf("Updated: %s\n", item.UpdatedAt)
}

// printTextDetails выводит детальную информацию записи типа text
func printTextDetails(item storage.Item) {
	fmt.Printf("Title: %s\n", item.Title)
	fmt.Printf("Type: %s\n", item.Type)
	fmt.Printf("ID: %s\n", item.ID)
	fmt.Println("Content:")
	fmt.Println(item.Content)
	fmt.Printf("Updated: %s\n", item.UpdatedAt)
}

// printCardDetails выводит детальную информацию записи типа card
func printCardDetails(item storage.Item) {
	fmt.Printf("Title: %s\n", item.Title)
	fmt.Printf("Type: %s\n", item.Type)
	fmt.Printf("ID: %s\n", item.ID)
	fmt.Printf("Number: %s\n", item.Number)
	fmt.Printf("Expiry: %s/%s\n", item.ExpiryMonth, item.ExpiryYear)
	fmt.Printf("CVV: %s\n", item.CVV)
	fmt.Printf("Updated: %s\n", item.UpdatedAt)
}

func init() {
	rootCmd.AddCommand(getCmd)
}