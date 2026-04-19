package cli

import (
	"fmt"
	"strings"
	"time"

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

		// Редактируем запись
		// Показываем текущие значения
		container.UI.Input.Printf("\nРедактирование записи: %s\n", item.Title)
		container.UI.Input.Printf("Тип: %s\n", item.Type)
		container.UI.Input.Printf("ID: %s\n\n", item.ID)

		// Редактирование названия
		container.UI.Input.Printf("Текущее название: %s\n", item.Title)
		container.UI.Input.Print("Введите новое название (оставьте пустым для сохранения текущего): ")
		newTitle, err := container.UI.Input.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ошибка чтения названия: %w", err)
		}
		newTitle = strings.TrimSpace(newTitle)
		if newTitle != "" {
			item.Title = newTitle
		}

		// Редактирование метаданных
		container.UI.Input.Printf("\nТекущие метаданные: %v\n", item.Metadata)
		container.UI.Input.Print("Введите новые метаданные (опционально, формат: ключ=значение, разделенные пробелом, оставьте пустым для сохранения текущих): ")
		metadataInput, _ := container.UI.Input.ReadString('\n')
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
			container.UI.Input.Printf("\nТекущий логин: %s\n", item.Username)
			container.UI.Input.Print("Введите новый логин (оставьте пустым для сохранения текущего): ")
			newUsername, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			newUsername = strings.TrimSpace(newUsername)
			if newUsername != "" {
				item.Username = newUsername
			}

			// Редактирование пароля
			container.UI.Input.Print("Введите новый пароль (оставьте пустым для сохранения текущего): ")
			newPasswordBytes, err := container.UI.Input.ReadPassword()
			container.UI.Input.Println("") // Добавляем перевод строки после ввода пароля
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			newPassword := string(newPasswordBytes)
			if newPassword != "" {
				item.Password = newPassword
			}

		case "text":
			// Редактирование содержимого заметки
			container.UI.Input.Printf("\nТекущее содержимое:\n%s\n", item.Content)
			container.UI.Input.Println("Введите новое содержимое заметки (завершите пустой строкой, оставьте пустым для сохранения текущего):")
			var contentLines []string
			for {
				line, err := container.UI.Input.ReadString('\n')
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
			container.UI.Input.Printf("\nТекущий номер карты: %s\n", item.Number)
			container.UI.Input.Print("Введите новый номер карты (оставьте пустым для сохранения текущего): ")
			newNumber, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения номера карты: %w", err)
			}
			newNumber = strings.TrimSpace(newNumber)
			if newNumber != "" {
				item.Number = newNumber
			}

			// Редактирование месяца истечения
			container.UI.Input.Printf("Текущий месяц истечения: %s\n", item.ExpiryMonth)
			container.UI.Input.Print("Введите новый месяц истечения (MM, оставьте пустым для сохранения текущего): ")
			newExpiryMonth, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения месяца: %w", err)
			}
			newExpiryMonth = strings.TrimSpace(newExpiryMonth)
			if newExpiryMonth != "" {
				item.ExpiryMonth = newExpiryMonth
			}

			// Редактирование года истечения
			container.UI.Input.Printf("Текущий год истечения: %s\n", item.ExpiryYear)
			container.UI.Input.Print("Введите новый год истечения (YY, оставьте пустым для сохранения текущего): ")
			newExpiryYear, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения года: %w", err)
			}
			newExpiryYear = strings.TrimSpace(newExpiryYear)
			if newExpiryYear != "" {
				item.ExpiryYear = newExpiryYear
			}

			// Редактирование CVV
			container.UI.Input.Printf("Текущий CVV: %s\n", item.CVV)
			container.UI.Input.Print("Введите новый CVV (оставьте пустым для сохранения текущего): ")
			newCVV, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения CVV: %w", err)
			}
			newCVV = strings.TrimSpace(newCVV)
			if newCVV != "" {
				item.CVV = newCVV
			}
		}

		// Обновляем время изменения записи
		item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

		// Обновляем секрет через сервис
		if err := container.SecretService.UpdateSecret(cmd.Context(), storageData, item); err != nil {
			container.UI.Output.Errorf("Ошибка обновления секрета: %v", err)
			return fmt.Errorf("ошибка обновления секрета: %w", err)
		}

		// Сохраняем секреты
		if err := container.SecretStorage.Save(storageData, session.Key); err != nil {
			container.UI.Output.Errorf("Ошибка сохранения секретов: %v", err)
			return fmt.Errorf("ошибка сохранения секретов: %w", err)
		}

		container.UI.Output.Success("Запись успешно обновлена.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
