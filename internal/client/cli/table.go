package cli

import (
	"fmt"
	"os"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// TableRenderer предоставляет функции для рендеринга таблиц
type TableRenderer struct{}

// NewTableRenderer создает новый рендерер таблиц
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{}
}

// RenderDuplicatesTable выводит таблицу найденных дубликатов
func (tr *TableRenderer) RenderDuplicatesTable(items []storage.Item) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal},
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: 30},
			},
		}),
	)
	table.Header("ID", "Title", "Type")

	var data [][]any
	for _, item := range items {
		id := item.ID
		if len(id) > 16 {
			id = id[:16]
		}
		data = append(data, []any{
			id,
			item.Title,
			item.Type,
		})
	}

	table.Bulk(data)
	table.Render()
}

// RenderLoginTable выводит таблицу записей типа login
func (tr *TableRenderer) RenderLoginTable(items []storage.Item) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal},
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: 30},
			},
		}),
	)
	table.Header("ID", "Title", "Username", "Password", "Updated")

	var data [][]any
	for _, item := range items {
		id := item.ID
		if len(id) > 8 {
			id = id[:8]
		}
		data = append(data, []any{
			id,
			item.Title,
			item.Username,
			"********", // Маскируем пароль
			item.UpdatedAt,
		})
	}

	table.Bulk(data)
	table.Render()
}

// RenderTextTable выводит таблицу записей типа text
func (tr *TableRenderer) RenderTextTable(items []storage.Item) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal},
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: 50},
			},
		}),
	)
	table.Header("ID", "Title", "Updated")

	var data [][]any
	for _, item := range items {
		id := item.ID
		if len(id) > 8 {
			id = id[:8]
		}
		data = append(data, []any{
			id,
			item.Title,
			item.UpdatedAt,
		})
	}

	table.Bulk(data)
	table.Render()
}

// RenderCardTable выводит таблицу записей типа card
func (tr *TableRenderer) RenderCardTable(items []storage.Item) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal},
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: 30},
			},
		}),
	)
	table.Header("ID", "Title", "Number", "Expiry", "CVV", "Updated")

	var data [][]any
	for _, item := range items {
		id := item.ID
		if len(id) > 8 {
			id = id[:8]
		}

		// Маскируем номер карты (показываем первые 4 и последние 4 цифры)
		maskedNumber := maskCardNumber(item.Number)

		// Формируем срок действия
		expiry := fmt.Sprintf("%s/%s", item.ExpiryMonth, item.ExpiryYear)

		data = append(data, []any{
			id,
			item.Title,
			maskedNumber,
			expiry,
			"***", // Маскируем CVV
			item.UpdatedAt,
		})
	}

	table.Bulk(data)
	table.Render()
}

// maskCardNumber маскирует середину номера карты
func maskCardNumber(number string) string {
	if len(number) < 8 {
		return number // Если номер слишком короткий, не маскируем
	}
	
	// Показываем первые 4 и последние 4 цифры
	prefix := number[:4]
	suffix := number[len(number)-4:]
	middle := "********"
	
	return prefix + middle + suffix
}