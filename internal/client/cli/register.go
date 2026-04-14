package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// registerCmd представляет команду регистрации
var registerCmd = &cobra.Command{
	Use:   "register [логин]",
	Short: "Регистрация нового пользователя",
	Long: `Регистрация нового пользователя в системе GophKeeper.
	
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

		// Подтверждение пароля
		fmt.Print("Подтвердите пароль: ")
		confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Добавляем перевод строки после ввода пароля
		if err != nil {
			return fmt.Errorf("ошибка чтения подтверждения пароля: %w", err)
		}
		confirmPassword := string(confirmPasswordBytes)

		if password != confirmPassword {
			output.Error("Пароли не совпадают")
			return fmt.Errorf("пароли не совпадают")
		}

		// Выполняем регистрацию
		output.Debugf("Регистрация пользователя: %s", login)

		err = authService.Register(cmd.Context(), login, password)
		if err != nil {
			output.Errorf("Ошибка регистрации: %v", err)
			return fmt.Errorf("ошибка регистрации: %w", err)
		}

		output.Success("Регистрация успешна!")
		output.Plain("Теперь вы можете войти с помощью команды: keeper login")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
