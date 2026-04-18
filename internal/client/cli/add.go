package cli

import (
	"fmt"
	"strings"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// addCmd представляет команду добавления новой записи
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Добавить новую запись",
	Long: `Добавить новую запись в хранилище секретов.
	
Поддерживаемые типы записей:
  - login: логин и пароль
  - text: текстовая заметка
  - card: данные банковской карты`,
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

		// Этап 2: Интерактивный ввод данных

		// Выбор типа записи
		container.UI.Input.Println("\nДоступные типы записей:")
		container.UI.Input.Println("  1. login - логин и пароль")
		container.UI.Input.Println("  2. text  - текстовая заметка")
		container.UI.Input.Println("  3. card  - данные банковской карты")
		container.UI.Input.Print("\nВыберите тип записи (1-3): ")

		typeChoice, err := container.UI.Input.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ошибка чтения типа: %w", err)
		}
		typeChoice = strings.TrimSpace(typeChoice)

		var itemType string
		switch typeChoice {
		case "1":
			itemType = "login"
		case "2":
			itemType = "text"
		case "3":
			itemType = "card"
		default:
			container.UI.Output.Error("Неверный выбор типа записи")
			return fmt.Errorf("неверный тип записи: %s", typeChoice)
		}

		// Ввод названия
		container.UI.Input.Print("Введите название записи: ")
		title, err := container.UI.Input.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ошибка чтения названия: %w", err)
		}
		title = strings.TrimSpace(title)

		if title == "" {
			container.UI.Output.Error("Название не может быть пустым")
			return fmt.Errorf("название не может быть пустым")
		}

		// Ввод метаданных (опционально)
		container.UI.Input.Print("Введите метаданные (опционально, формат: ключ=значение, разделенные пробелом): ")
		metadataInput, _ := container.UI.Input.ReadString('\n')
		metadataInput = strings.TrimSpace(metadataInput)

		metadata := make(map[string]string)
		if metadataInput != "" {
			parts := strings.Fields(metadataInput)
			for _, part := range parts {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					metadata[kv[0]] = kv[1]
				}
			}
		}

		// Создаем новую запись
		item := storage.NewItem(itemType, title)
		item.Metadata = metadata

		// Специфичные поля для каждого типа
		switch itemType {
		case "login":
			container.UI.Input.Print("Введите логин: ")
			username, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			item.Username = strings.TrimSpace(username)

			container.UI.Input.Print("Введите пароль: ")
			passwordBytes, err := container.UI.Input.ReadPassword()
			container.UI.Input.Println("") // Добавляем перевод строки после ввода пароля
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			item.Password = string(passwordBytes)

		case "text":
			container.UI.Input.Println("Введите содержимое заметки (завершите пустой строкой):")
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
			item.Content = strings.Join(contentLines, "\n")

		case "card":
			container.UI.Input.Print("Введите номер карты: ")
			number, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения номера карты: %w", err)
			}
			item.Number = strings.TrimSpace(number)

			container.UI.Input.Print("Введите месяц истечения (MM): ")
			expiryMonth, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения месяца: %w", err)
			}
			item.ExpiryMonth = strings.TrimSpace(expiryMonth)

			container.UI.Input.Print("Введите год истечения (YY): ")
			expiryYear, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения года: %w", err)
			}
			item.ExpiryYear = strings.TrimSpace(expiryYear)

			container.UI.Input.Print("Введите CVV: ")
			cvv, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения CVV: %w", err)
			}
			item.CVV = strings.TrimSpace(cvv)
		}

		// Этап 3: Работа с файлом секретов

		// Загружаем существующие секреты или создаем новые
		storageData, err := container.SecretStorage.Load(session.Key)
		if err != nil {
			container.UI.Output.Errorf("Ошибка загрузки секретов: %v", err)
			return fmt.Errorf("ошибка загрузки секретов: %w", err)
		}

		// Этап 4: Добавление секрета через сервис
		if err := container.SecretService.AddSecret(cmd.Context(), storageData, &item); err != nil {
			container.UI.Output.Errorf("Ошибка добавления секрета: %v", err)
			return fmt.Errorf("ошибка добавления секрета: %w", err)
		}

		// Этап 5: Сохранение секретов
		if err := container.SecretStorage.Save(storageData, session.Key); err != nil {
			container.UI.Output.Errorf("Ошибка сохранения секретов: %v", err)
			return fmt.Errorf("ошибка сохранения секретов: %w", err)
		}

		container.UI.Output.Success("Секрет успешно добавлен и сохранен.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
