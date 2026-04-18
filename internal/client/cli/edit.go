package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
	"github.com/spf13/cobra"
)

// editCmd представляет команду редактирования записи
var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Редактировать запись",
	Long: `Редактировать существующую запись в хранилище секретов.
	
Команда находит запись по началу ID и позволяет интерактивно редактировать её поля.
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
			// Редактируем запись
			item := &storageData.Items[foundIndex]

			// Показываем текущие значения
			uiInstance.Input.Printf("\nРедактирование записи: %s\n", item.Title)
			uiInstance.Input.Printf("Тип: %s\n", item.Type)
			uiInstance.Input.Printf("ID: %s\n\n", item.ID)

			// Редактирование названия
			uiInstance.Input.Printf("Текущее название: %s\n", item.Title)
			uiInstance.Input.Print("Введите новое название (оставьте пустым для сохранения текущего): ")
			newTitle, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения названия: %w", err)
			}
			newTitle = strings.TrimSpace(newTitle)
			if newTitle != "" {
				item.Title = newTitle
			}

			// Редактирование метаданных
			uiInstance.Input.Printf("\nТекущие метаданные: %v\n", item.Metadata)
			uiInstance.Input.Print("Введите новые метаданные (опционально, формат: ключ=значение, разделенные пробелом, оставьте пустым для сохранения текущих): ")
			metadataInput, _ := uiInstance.Input.ReadString('\n')
			metadataInput = strings.TrimSpace(metadataInput)

			if metadataInput != "" {
				newMetadata := make(map[string]string)
				parts := strings.Fields(metadataInput)
				for _, part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						newMetadata[kv[0]] = kv[1]
					}
				}
				item.Metadata = newMetadata
			}

			// Специфичные поля для каждого типа
			switch item.Type {
			case "login":
				// Редактирование логина
				uiInstance.Input.Printf("\nТекущий логин: %s\n", item.Username)
				uiInstance.Input.Print("Введите новый логин (оставьте пустым для сохранения текущего): ")
				newUsername, err := uiInstance.Input.ReadString('\n')
				if err != nil {
					return fmt.Errorf("ошибка чтения логина: %w", err)
				}
				newUsername = strings.TrimSpace(newUsername)
				if newUsername != "" {
					item.Username = newUsername
				}

				// Редактирование пароля
				uiInstance.Input.Print("Введите новый пароль (оставьте пустым для сохранения текущего): ")
				newPasswordBytes, err := uiInstance.Input.ReadPassword()
				uiInstance.Input.Println("") // Добавляем перевод строки после ввода пароля
				if err != nil {
					return fmt.Errorf("ошибка чтения пароля: %w", err)
				}
				newPassword := string(newPasswordBytes)
				if newPassword != "" {
					item.Password = newPassword
				}

			case "text":
				// Редактирование содержимого заметки
				uiInstance.Input.Printf("\nТекущее содержимое:\n%s\n", item.Content)
				uiInstance.Input.Println("Введите новое содержимое заметки (завершите пустой строкой, оставьте пустым для сохранения текущего):")
				var contentLines []string
				for {
					line, err := uiInstance.Input.ReadString('\n')
					if err != nil {
						return fmt.Errorf("ошибка чтения содержимого: %w", err)
					}
					line = strings.TrimRight(line, "\n")
					if line == "" {
						break
					}
					contentLines = append(contentLines, line)
				}
				if len(contentLines) > 0 {
					item.Content = strings.Join(contentLines, "\n")
				}

			case "card":
				// Редактирование номера карты
				uiInstance.Input.Printf("\nТекущий номер карты: %s\n", item.Number)
				uiInstance.Input.Print("Введите новый номер карты (оставьте пустым для сохранения текущего): ")
				newNumber, err := uiInstance.Input.ReadString('\n')
				if err != nil {
					return fmt.Errorf("ошибка чтения номера карты: %w", err)
				}
				newNumber = strings.TrimSpace(newNumber)
				if newNumber != "" {
					item.Number = newNumber
				}

				// Редактирование месяца истечения
				uiInstance.Input.Printf("Текущий месяц истечения: %s\n", item.ExpiryMonth)
				uiInstance.Input.Print("Введите новый месяц истечения (MM, оставьте пустым для сохранения текущего): ")
				newExpiryMonth, err := uiInstance.Input.ReadString('\n')
				if err != nil {
					return fmt.Errorf("ошибка чтения месяца: %w", err)
				}
				newExpiryMonth = strings.TrimSpace(newExpiryMonth)
				if newExpiryMonth != "" {
					item.ExpiryMonth = newExpiryMonth
				}

				// Редактирование года истечения
				uiInstance.Input.Printf("Текущий год истечения: %s\n", item.ExpiryYear)
				uiInstance.Input.Print("Введите новый год истечения (YY, оставьте пустым для сохранения текущего): ")
				newExpiryYear, err := uiInstance.Input.ReadString('\n')
				if err != nil {
					return fmt.Errorf("ошибка чтения года: %w", err)
				}
				newExpiryYear = strings.TrimSpace(newExpiryYear)
				if newExpiryYear != "" {
					item.ExpiryYear = newExpiryYear
				}

				// Редактирование CVV
				uiInstance.Input.Printf("Текущий CVV: %s\n", item.CVV)
				uiInstance.Input.Print("Введите новый CVV (оставьте пустым для сохранения текущего): ")
				newCVV, err := uiInstance.Input.ReadString('\n')
				if err != nil {
					return fmt.Errorf("ошибка чтения CVV: %w", err)
				}
				newCVV = strings.TrimSpace(newCVV)
				if newCVV != "" {
					item.CVV = newCVV
				}
			}

			// Обновляем время изменения
			item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			storageData.LastModified = time.Now().UTC().Format(time.RFC3339)

			// Сохраняем данные
			if err := dataStorage.Save(storageData, session.Key); err != nil {
				uiInstance.Output.Errorf("Ошибка сохранения данных: %v", err)
				return fmt.Errorf("ошибка сохранения данных: %w", err)
			}

			uiInstance.Output.Success("Запись успешно обновлена.")

		default:
			// Найдено несколько записей - выводим таблицу дубликатов
			uiInstance.Output.Plain(fmt.Sprintf("Найдено несколько записей с префиксом '%s', уточните запрос:", idPrefix))
			uiInstance.TableRenderer.RenderDuplicatesTable(foundItems)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
