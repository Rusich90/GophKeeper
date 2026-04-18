package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
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

		// Валидация типа данных
		dataType := args[0]
		if dataType != "login" && dataType != "text" && dataType != "card" {
			output.Error("Неизвестный тип данных. Доступные: login, text, card")
			return fmt.Errorf("неизвестный тип данных: %s", dataType)
		}

		// Получаем ключ шифрования из сессии
		session, err := sessionStorage.LoadSession()
		if err != nil {
			output.Error("Сессия не инициализирована. Пожалуйста, войдите в систему.")
			return fmt.Errorf("сессия не инициализирована: %w", err)
		}

		// Загружаем данные
		dataStorage := storage.NewDataStorage()
		storageData, err := dataStorage.Load(session.Key)
		if err != nil {
			output.Errorf("Ошибка загрузки данных: %v", err)
			return fmt.Errorf("ошибка загрузки данных: %w", err)
		}

		// Фильтруем записи по типу
		var filteredItems []storage.Item
		for _, item := range storageData.Items {
			if item.Type == dataType {
				filteredItems = append(filteredItems, item)
			}
		}

		// Проверяем, есть ли записи
		if len(filteredItems) == 0 {
			output.Plain("Записи не найдены")
			return nil
		}

		// Получаем рендерер таблиц из контекста
		tableRenderer, ok := cmd.Context().Value("tableRenderer").(*TableRenderer)
		if !ok {
			return fmt.Errorf("рендерер таблиц не инициализирован")
		}

		// Диспетчеризация функций рендеринга
		switch dataType {
		case "login":
			tableRenderer.RenderLoginTable(filteredItems)
		case "text":
			tableRenderer.RenderTextTable(filteredItems)
		case "card":
			tableRenderer.RenderCardTable(filteredItems)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}