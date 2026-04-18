package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
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
		var foundIndex = -1
		
		for i, item := range storageData.Items {
			if strings.HasPrefix(item.ID, idPrefix) {
				foundItems = append(foundItems, item)
				if foundIndex == -1 {
					foundIndex = i
				}
			}
		}

		// Обработка результатов поиска
		switch len(foundItems) {
		case 0:
			uiInstance.Output.Errorf("Запись с ID '%s' не найдена", idPrefix)
			return fmt.Errorf("запись не найдена")
			
		case 1:
			// Удаляем запись
			item := foundItems[0]
			// Удаляем элемент из среза
			storageData.Items = append(storageData.Items[:foundIndex], storageData.Items[foundIndex+1:]...)
			
			// Обновляем LastModified всей структуры
			storageData.LastModified = time.Now().UTC().Format(time.RFC3339)
			
			// Сохраняем данные
			if err := dataStorage.Save(storageData, session.Key); err != nil {
				uiInstance.Output.Errorf("Ошибка сохранения данных: %v", err)
				return fmt.Errorf("ошибка сохранения данных: %w", err)
			}
			
			uiInstance.Output.Success(fmt.Sprintf("Запись '%s' успешно удалена", item.Title))
			
		default:
			// Найдено несколько записей - выводим таблицу дубликатов
			uiInstance.Output.Plain(fmt.Sprintf("Найдено несколько записей с префиксом '%s', уточните запрос:", idPrefix))
			uiInstance.TableRenderer.RenderDuplicatesTable(foundItems)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}