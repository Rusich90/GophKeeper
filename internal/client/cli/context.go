package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// containerKey - типизированный ключ для хранения контейнера в контексте
type containerKey struct{}

// GetContainer извлекает контейнер зависимостей из контекста команды
// Возвращает ошибку, если контейнер не инициализирован
func GetContainer(cmd *cobra.Command) (*Container, error) {
	container, ok := cmd.Context().Value(containerKey{}).(*Container)
	if !ok {
		return nil, fmt.Errorf("контейнер не инициализирован")
	}
	return container, nil
}

// LoadSession загружает сессию из хранилища
// Возвращает ошибку, если сессия не инициализирована
func LoadSession(container *Container) (*storage.Session, error) {
	session, err := container.SessionStorage.LoadSession()
	if err != nil {
		container.UI.Output.Error("Сессия не инициализирована. Пожалуйста, войдите в систему.")
		return nil, fmt.Errorf("сессия не инициализирована: %w", err)
	}
	return session, nil
}

// LoadSessionOptional загружает сессию из хранилища, но не возвращает ошибку, если сессия отсутствует
// Полезно для команд, которые могут работать без авторизации
func LoadSessionOptional(container *Container) (*storage.Session, error) {
	session, err := container.SessionStorage.LoadSession()
	if err != nil {
		// Сессия отсутствует, но это не ошибка
		return nil, nil
	}
	return session, nil
}
