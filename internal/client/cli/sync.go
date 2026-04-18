package cli

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Синхронизировать данные с сервером",
	Long:  `Синхронизирует локальные данные с сервером, используя стратегию объединения изменений (Merge) на основе временных меток.`,
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Получаем контейнер из контекста
	container, err := GetContainer(cmd)
	if err != nil {
		return err
	}

	// Загружаем сессию
	session, err := LoadSession(container)
	if err != nil {
		return err
	}

	// Выполняем синхронизацию
	if err := container.SyncService.Sync(cmd.Context(), session); err != nil {
		container.UI.Output.Errorf("Ошибка синхронизации: %v", err)
		return err
	}

	container.UI.Output.Success("Синхронизация завершена успешно")
	return nil
}
