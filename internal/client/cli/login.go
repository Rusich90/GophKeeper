package cli

import (
	"fmt"
	"strings"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
	"github.com/spf13/cobra"
)

// loginCmd представляет команду входа
var loginCmd = &cobra.Command{
	Use:   "login [логин]",
	Short: "Авторизация пользователя",
	Long: `Авторизация пользователя в системе GophKeeper.
	
Пароль запрашивается интерактивно для безопасности.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем UI из контекста
		uiInstance, ok := cmd.Context().Value("ui").(*ui.UI)
		if !ok {
			return fmt.Errorf("UI не инициализирован")
		}

		// Получаем сервис авторизации из контекста
		authService, ok := cmd.Context().Value("authService").(service.AuthService)
		if !ok {
			return fmt.Errorf("сервис авторизации не инициализирован")
		}

		// Получаем хранилище сессий из контекста
		sessionStorage, ok := cmd.Context().Value("sessionStorage").(*storage.SessionStorage)
		if !ok {
			return fmt.Errorf("хранилище сессий не инициализировано")
		}

		// Получаем логин
		var login string
		if len(args) > 0 {
			login = strings.TrimSpace(args[0])
		} else {
			uiInstance.Input.Print("Введите логин: ")
			loginInput, err := uiInstance.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			login = strings.TrimSpace(loginInput)
		}

		if login == "" {
			uiInstance.Output.Error("Логин не может быть пустым")
			return fmt.Errorf("логин не может быть пустым")
		}

		// Получаем пароль интерактивно
		uiInstance.Input.Print("Введите пароль: ")
		passwordBytes, err := uiInstance.Input.ReadPassword()
		uiInstance.Input.Println("") // Добавляем перевод строки после ввода пароля
		if err != nil {
			return fmt.Errorf("ошибка чтения пароля: %w", err)
		}
		password := string(passwordBytes)

		if password == "" {
			uiInstance.Output.Error("Пароль не может быть пустым")
			return fmt.Errorf("пароль не может быть пустым")
		}

		// Выполняем вход
		uiInstance.Output.Debugf("Вход пользователя: %s", login)

		token, err := authService.Login(cmd.Context(), login, password)
		if err != nil {
			uiInstance.Output.Errorf("Ошибка входа: %v", err)
			return fmt.Errorf("ошибка входа: %w", err)
		}

		// Генерируем ключ шифрования из пароля
		encryptionKey := crypto.GenerateEncryptionKey(password)

		// Создаем сессию с токеном и ключом шифрования
		session := &storage.Session{
			Token: token,
			Key:   encryptionKey,
		}

		// Сохраняем сессию
		if err := sessionStorage.SaveSession(session); err != nil {
			uiInstance.Output.Errorf("Ошибка сохранения сессии: %v", err)
			return fmt.Errorf("ошибка сохранения сессии: %w", err)
		}

		uiInstance.Output.Success("Вход выполнен успешно!")
		uiInstance.Output.Plain("Сессия сохранена. Вы авторизованы.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
