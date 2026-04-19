package cli

import (
	"fmt"
	"strings"

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

		// Подтверждение пароля
		container.UI.Input.Print("Подтвердите пароль: ")
		confirmPasswordBytes, err := container.UI.Input.ReadPassword()
		container.UI.Input.Println("") // Добавляем перевод строки после ввода пароля
		if err != nil {
			return fmt.Errorf("ошибка чтения подтверждения пароля: %w", err)
		}
		confirmPassword := string(confirmPasswordBytes)

		if password != confirmPassword {
			container.UI.Output.Error("Пароли не совпадают")
			return fmt.Errorf("пароли не совпадают")
		}

		// Выполняем регистрацию
		container.UI.Output.Debugf("Регистрация пользователя: %s", login)

		err = container.AuthService.Register(cmd.Context(), login, password)
		if err != nil {
			container.UI.Output.Errorf("Ошибка регистрации: %v", err)
			return fmt.Errorf("ошибка регистрации: %w", err)
		}

		container.UI.Output.Success("Регистрация успешна!")
		container.UI.Output.Plain("Теперь вы можете войти с помощью команды: keeper login")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
