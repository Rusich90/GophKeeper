package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// loginCmd представляет команду входа
var loginCmd = &cobra.Command{
	Use:   "login [логин]",
	Short: "Авторизация пользователя",
	Long: `Авторизация пользователя в системе GophKeeper.
	
Пароль запрашивается интерактивно для безопасности.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Получаем output из контекста
		output, ok := cmd.Context().Value("output").(*ColorOutput)
		if !ok {
			return fmt.Errorf("сервис вывода не инициализирован")
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
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Введите логин: ")
			loginInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}
			login = strings.TrimSpace(loginInput)
		}

		if login == "" {
			output.Error("Логин не может быть пустым")
			return fmt.Errorf("логин не может быть пустым")
		}

		// Получаем пароль интерактивно
		fmt.Print("Введите пароль: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Добавляем перевод строки после ввода пароля
		if err != nil {
			return fmt.Errorf("ошибка чтения пароля: %w", err)
		}
		password := string(passwordBytes)

		if password == "" {
			output.Error("Пароль не может быть пустым")
			return fmt.Errorf("пароль не может быть пустым")
		}

		// Выполняем вход
		output.Debugf("Вход пользователя: %s", login)

		token, err := authService.Login(cmd.Context(), login, password)
		if err != nil {
			output.Errorf("Ошибка входа: %v", err)
			return fmt.Errorf("ошибка входа: %w", err)
		}

		// Генерируем ключ шифрования из пароля
		encryptionKey := storage.GenerateEncryptionKey(password)

		// Создаем сессию с токеном и ключом шифрования
		session := &storage.Session{
			Token: token,
			Key:   encryptionKey,
		}

		// Сохраняем сессию
		if err := sessionStorage.SaveSession(session); err != nil {
			output.Errorf("Ошибка сохранения сессии: %v", err)
			return fmt.Errorf("ошибка сохранения сессии: %w", err)
		}

		output.Success("Вход выполнен успешно!")
		output.Plain("Сессия сохранена. Вы авторизованы.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
