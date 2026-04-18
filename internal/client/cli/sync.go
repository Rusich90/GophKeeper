package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client/grpc"
	"github.com/Rusich90/GophKeeper/internal/client/service"
	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/Rusich90/GophKeeper/internal/client/ui"
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
	ctx := cmd.Context()
	uiInstance := ctx.Value("ui").(*ui.UI)
	grpcClient := ctx.Value("grpcClient").(*grpc.Client)
	sessionStorage := ctx.Value("sessionStorage").(*storage.SessionStorage)

	// Загружаем сессию
	session, err := sessionStorage.LoadSession()
	if err != nil {
		uiInstance.Output.Errorf("Ошибка загрузки сессии: %v", err)
		return fmt.Errorf("ошибка загрузки сессии: %w", err)
	}

	// Создаем сервис синхронизации
	syncService := service.NewSyncService(grpcClient, sessionStorage)

	// Выполняем синхронизацию
	if err := syncService.Sync(ctx, session); err != nil {
		uiInstance.Output.Errorf("Ошибка синхронизации: %v", err)
		return err
	}

	uiInstance.Output.Success("Синхронизация завершена успешно")
	return nil
}
