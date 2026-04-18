package cli

import (
	"fmt"
	"strings"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
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

		// Этап 1: Получение ключа шифрования из сессии
		session, err := sessionStorage.LoadSession()
		if err != nil {
			uiInstance.Output.Error("Сессия не инициализирована. Пожалуйста, войдите в систему.")
			return fmt.Errorf("сессия не инициализирована: %w", err)
		}

		// Этап 2: Интерактивный ввод данных

		// Выбор типа записи
		uiInstance.Input.Println("\nДоступные типы записей:")
		uiInstance.Input.Println("  1. login - логин и пароль")
		uiInstance.Input.Println("  2. text  - текстовая заметка")
		uiInstance.Input.Println("  3. card  - данные банковской карты")
		uiInstance.Input.Print("\nВыберите тип записи (1-3): ")

		typeChoice, err := uiInstance.Input.ReadString('\n')
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
			uiInstance.Output.Error("Неверный выбор типа записи")
			return fmt.Errorf("неверный тип записи: %s", typeChoice)
		}

		// Ввод названия
		uiInstance.Input.Print("Введите название записи: ")
		title, err := uiInstance.Input.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ошибка чтения названия: %w", err)
		}
		title = strings.TrimSpace(title)

		if title == "" {
			uiInstance.Output.Error("Название не может быть пустым")
			return fmt.Errorf("название не может быть пустым")
		}

		// Ввод метаданных (опционально)
		uiInstance.Input.Print("Введите метаданные (опционально, формат: ключ=значение, разделенные пробелом): ")
		metadataInput, _ := uiInstance.Input.ReadString('\n')
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
			uiInstance.Input.Print("Введите логин: ")
			username, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			item.Username = strings.TrimSpace(username)

			uiInstance.Input.Print("Введите пароль: ")
			passwordBytes, err := uiInstance.Input.ReadPassword()
			uiInstance.Input.Println("") // Добавляем перевод строки после ввода пароля
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			item.Password = string(passwordBytes)

		case "text":
			uiInstance.Input.Println("Введите содержимое заметки (завершите пустой строкой):")
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
			item.Content = strings.Join(contentLines, "\n")

		case "card":
			uiInstance.Input.Print("Введите номер карты: ")
			number, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения номера карты: %w", err)
			}
			item.Number = strings.TrimSpace(number)

			uiInstance.Input.Print("Введите месяц истечения (MM): ")
			expiryMonth, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения месяца: %w", err)
			}
			item.ExpiryMonth = strings.TrimSpace(expiryMonth)

			uiInstance.Input.Print("Введите год истечения (YY): ")
			expiryYear, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения года: %w", err)
			}
			item.ExpiryYear = strings.TrimSpace(expiryYear)

			uiInstance.Input.Print("Введите CVV: ")
			cvv, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения CVV: %w", err)
			}
			item.CVV = strings.TrimSpace(cvv)
		}

		// Этап 3: Работа с файлом данных
		dataStorage := storage.NewDataStorage()

		// Загружаем существующие данные или создаем новые
		storageData, err := dataStorage.Load(session.Key)
		if err != nil {
			uiInstance.Output.Errorf("Ошибка загрузки данных: %v", err)
			return fmt.Errorf("ошибка загрузки данных: %w", err)
		}

		// Этап 4: Модификация данных
		storageData.Items = append(storageData.Items, item)

		// Этап 5: Сохранение данных
		if err := dataStorage.Save(storageData, session.Key); err != nil {
			uiInstance.Output.Errorf("Ошибка сохранения данных: %v", err)
			return fmt.Errorf("ошибка сохранения данных: %w", err)
		}

		uiInstance.Output.Success("Запись успешно добавлена и сохранена.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
