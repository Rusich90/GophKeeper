package cli

import (
	"fmt"
	"strings"

	"github.com/Rusich90/GophKeeper/internal/client/crypto"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
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
		// Получаем контейнер из контекста
		container, err := GetContainer(cmd)
		if err != nil {
			return err
		}

		// Получаем логин
		var login string
		if len(args) > 0 {
			login = strings.TrimSpace(args[0])
		} else {
			container.UI.Input.Print("Введите логин: ")
			loginInput, err := container.UI.Input.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			login = strings.TrimSpace(loginInput)
		}

		if login == "" {
			container.UI.Output.Error("Логин не может быть пустым")
			return fmt.Errorf("логин не может быть пустым")
		}

		// Получаем пароль интерактивно
		container.UI.Input.Print("Введите пароль: ")
		passwordBytes, err := container.UI.Input.ReadPassword()
		container.UI.Input.Println("") // Добавляем перевод строки после ввода пароля
		if err != nil {
			return fmt.Errorf("ошибка чтения пароля: %w", err)
		}
		password := string(passwordBytes)

		if password == "" {
			container.UI.Output.Error("Пароль не может быть пустым")
			return fmt.Errorf("пароль не может быть пустым")
		}

		// Выполняем вход
		container.UI.Output.Debugf("Вход пользователя: %s", login)

		token, err := container.AuthService.Login(cmd.Context(), login, password)
		if err != nil {
			container.UI.Output.Errorf("Ошибка входа: %v", err)
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
		if err := container.SessionStorage.SaveSession(session); err != nil {
			container.UI.Output.Errorf("Ошибка сохранения сессии: %v", err)
			return fmt.Errorf("ошибка сохранения сессии: %w", err)
		}

		container.UI.Output.Success("Вход выполнен успешно!")
		container.UI.Output.Plain("Сессия сохранена. Вы авторизованы.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
