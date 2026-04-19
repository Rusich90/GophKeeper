package cli

import (
	"fmt"

	"github.com/Rusich90/GophKeeper/internal/client"
	"github.com/spf13/cobra"
)

// versionCmd представляет команду для отображения версии
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать информацию о версии клиента",
	Long:  `Отображает информацию о версии и дате сборки клиента GophKeeper.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GophKeeper CLI Client\n")
		fmt.Printf("Версия: %s\n", client.Version)
		fmt.Printf("Дата сборки: %s\n", client.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}