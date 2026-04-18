package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
		// Получаем output из контекста
		output, ok := cmd.Context().Value("output").(*ColorOutput)
		if !ok {
			return fmt.Errorf("сервис вывода не инициализирован")
		}

		// Получаем хранилище сессий из контекста
		sessionStorage, ok := cmd.Context().Value("sessionStorage").(*storage.SessionStorage)
		if !ok {
			return fmt.Errorf("хранилище сессий не инициализировано")
		}

		// Этап 1: Получение ключа шифрования из сессии
		session, err := sessionStorage.LoadSession()
		if err != nil {
			output.Error("Сессия не инициализирована. Пожалуйста, войдите в систему.")
			return fmt.Errorf("сессия не инициализирована: %w", err)
		}

		// Этап 2: Интерактивный ввод данных
		reader := bufio.NewReader(os.Stdin)

		// Выбор типа записи
		fmt.Println("\nДоступные типы записей:")
		fmt.Println("  1. login - логин и пароль")
		fmt.Println("  2. text  - текстовая заметка")
		fmt.Println("  3. card  - данные банковской карты")
		fmt.Print("\nВыберите тип записи (1-3): ")

		typeChoice, err := reader.ReadString('\n')
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
			output.Error("Неверный выбор типа записи")
			return fmt.Errorf("неверный тип записи: %s", typeChoice)
		}

		// Ввод названия
		fmt.Print("Введите название записи: ")
		title, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ошибка чтения названия: %w", err)
		}
		title = strings.TrimSpace(title)

		if title == "" {
			output.Error("Название не может быть пустым")
			return fmt.Errorf("название не может быть пустым")
		}

		// Ввод метаданных (опционально)
		fmt.Print("Введите метаданные (опционально, формат: ключ=значение, разделенные пробелом): ")
		metadataInput, _ := reader.ReadString('\n')
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
			fmt.Print("Введите логин: ")
			username, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			item.Username = strings.TrimSpace(username)

			fmt.Print("Введите пароль: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println() // Добавляем перевод строки после ввода пароля
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			item.Password = string(passwordBytes)

		case "text":
			fmt.Println("Введите содержимое заметки (завершите пустой строкой):")
			var contentLines []string
			for {
				line, err := reader.ReadString('\n')
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
			fmt.Print("Введите номер карты: ")
			number, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения номера карты: %w", err)
			}
			item.Number = strings.TrimSpace(number)

			fmt.Print("Введите месяц истечения (MM): ")
			expiryMonth, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения месяца: %w", err)
			}
			item.ExpiryMonth = strings.TrimSpace(expiryMonth)

			fmt.Print("Введите год истечения (YY): ")
			expiryYear, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения года: %w", err)
			}
			item.ExpiryYear = strings.TrimSpace(expiryYear)

			fmt.Print("Введите CVV: ")
			cvv, err := reader.ReadString('\n')
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
			output.Errorf("Ошибка загрузки данных: %v", err)
			return fmt.Errorf("ошибка загрузки данных: %w", err)
		}

		// Этап 4: Модификация данных
		storageData.Items = append(storageData.Items, item)

		// Этап 5: Сохранение данных
		if err := dataStorage.Save(storageData, session.Key); err != nil {
			output.Errorf("Ошибка сохранения данных: %v", err)
			return fmt.Errorf("ошибка сохранения данных: %w", err)
		}

		output.Success("Запись успешно добавлена и сохранена.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}