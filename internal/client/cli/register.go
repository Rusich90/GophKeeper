package cli

import (
	"fmt"
	"strings"

	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
	"github.com/spf13/cobra"
)

// registerCmd представляет команду регистрации
var registerCmd = &cobra.Command{
	Use:   "register [логин]",
	Short: "Регистрация нового пользователя",
	Long: `Регистрация нового пользователя в системе GophKeeper.
	
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

		// Подтверждение пароля
		uiInstance.Input.Print("Подтвердите пароль: ")
		confirmPasswordBytes, err := uiInstance.Input.ReadPassword()
		uiInstance.Input.Println("") // Добавляем перевод строки после ввода пароля
		if err != nil {
			return fmt.Errorf("ошибка чтения подтверждения пароля: %w", err)
		}
		confirmPassword := string(confirmPasswordBytes)

		if password != confirmPassword {
			uiInstance.Output.Error("Пароли не совпадают")
			return fmt.Errorf("пароли не совпадают")
		}

		// Выполняем регистрацию
		uiInstance.Output.Debugf("Регистрация пользователя: %s", login)

		err = authService.Register(cmd.Context(), login, password)
		if err != nil {
			uiInstance.Output.Errorf("Ошибка регистрации: %v", err)
			return fmt.Errorf("ошибка регистрации: %w", err)
		}

		uiInstance.Output.Success("Регистрация успешна!")
		uiInstance.Output.Plain("Теперь вы можете войти с помощью команды: keeper login")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
