package ui

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/Rusich90/GophKeeper/internal/client/storage"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"golang.org/x/term"
)

// Output интерфейс для вывода сообщений
type Output interface {
	Success(message string)
	Successf(format string, args ...interface{})
	Error(message string)
	Errorf(format string, args ...interface{})
	Warning(message string)
	Warningf(format string, args ...interface{})
	Info(message string)
	Infof(format string, args ...interface{})
	Plain(message string)
	Plainf(format string, args ...interface{})
	Debug(message string)
	Debugf(format string, args ...interface{})
	GetLogger() *slog.Logger
}

// Input интерфейс для ввода данных
type Input interface {
	Print(message string)
	Println(message string)
	Printf(format string, args ...interface{})
	ReadString(delim byte) (string, error)
	ReadPassword() ([]byte, error)
}

// TableRenderer интерфейс для рендеринга таблиц
type TableRenderer interface {
	RenderDuplicatesTable(items []storage.Item)
	RenderLoginTable(items []storage.Item)
	RenderTextTable(items []storage.Item)
	RenderCardTable(items []storage.Item)
}

// ConsoleOutput реализует Output интерфейс
type ConsoleOutput struct {
	logger *slog.Logger
}

// NewConsoleOutput создает новый консольный вывод
func NewConsoleOutput(verbose bool) *ConsoleOutput {
	level := slog.LevelError
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &ConsoleOutput{logger: logger}
}

// Success выводит успешное сообщение зеленым цветом
func (c *ConsoleOutput) Success(message string) {
	fmt.Printf("\033[32m✓ %s\033[0m\n", message)
}

// Successf выводит успешное сообщение с форматированием
func (c *ConsoleOutput) Successf(format string, args ...interface{}) {
	fmt.Printf("\033[32m✓ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Error выводит сообщение об ошибке красным цветом
func (c *ConsoleOutput) Error(message string) {
	fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", message)
}

// Errorf выводит сообщение об ошибке с форматированием
func (c *ConsoleOutput) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Warning выводит предупреждение желтым цветом
func (c *ConsoleOutput) Warning(message string) {
	fmt.Printf("\033[33m⚠ %s\033[0m\n", message)
}

// Warningf выводит предупреждение с форматированием
func (c *ConsoleOutput) Warningf(format string, args ...interface{}) {
	fmt.Printf("\033[33m⚠ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Info выводит информационное сообщение синим цветом
func (c *ConsoleOutput) Info(message string) {
	fmt.Printf("\033[34mℹ %s\033[0m\n", message)
}

// Infof выводит информационное сообщение с форматированием
func (c *ConsoleOutput) Infof(format string, args ...interface{}) {
	fmt.Printf("\033[34mℹ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Plain выводит обычное сообщение без цвета
func (c *ConsoleOutput) Plain(message string) {
	fmt.Println(message)
}

// Plainf выводит обычное сообщение с форматированием
func (c *ConsoleOutput) Plainf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Debug выводит отладочное сообщение (только в verbose режиме)
func (c *ConsoleOutput) Debug(message string) {
	c.logger.Debug(message)
}

// Debugf выводит отладочное сообщение с форматированием
func (c *ConsoleOutput) Debugf(format string, args ...interface{}) {
	c.logger.Debug(fmt.Sprintf(format, args...))
}

// GetLogger возвращает логгер
func (c *ConsoleOutput) GetLogger() *slog.Logger {
	return c.logger
}

// ConsoleInput реализует Input интерфейс
type ConsoleInput struct {
	reader *bufio.Reader
}

// NewConsoleInput создает новый консольный ввод
func NewConsoleInput() *ConsoleInput {
	return &ConsoleInput{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Print выводит сообщение без переноса строки
func (c *ConsoleInput) Print(message string) {
	fmt.Print(message)
}

// Println выводит сообщение с переносом строки
func (c *ConsoleInput) Println(message string) {
	fmt.Println(message)
}

// Printf выводит форматированное сообщение
func (c *ConsoleInput) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// ReadString читает строку до разделителя
func (c *ConsoleInput) ReadString(delim byte) (string, error) {
	return c.reader.ReadString(delim)
}

// ReadPassword читает пароль без отображения на экране
func (c *ConsoleInput) ReadPassword() ([]byte, error) {
	return term.ReadPassword(int(syscall.Stdin))
}

// ConsoleTableRenderer реализует TableRenderer интерфейс
type ConsoleTableRenderer struct{}

// NewConsoleTableRenderer создает новый рендерер таблиц
func NewConsoleTableRenderer() *ConsoleTableRenderer {
	return &ConsoleTableRenderer{}
}

// RenderDuplicatesTable выводит таблицу найденных дубликатов
func (tr *ConsoleTableRenderer) RenderDuplicatesTable(items []storage.Item) {
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
func (tr *ConsoleTableRenderer) RenderLoginTable(items []storage.Item) {
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
func (tr *ConsoleTableRenderer) RenderTextTable(items []storage.Item) {
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
func (tr *ConsoleTableRenderer) RenderCardTable(items []storage.Item) {
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

// UI объединяет все UI компоненты
type UI struct {
	Output        Output
	Input         Input
	TableRenderer TableRenderer
}

// NewUI создает новый UI
func NewUI(verbose bool) *UI {
	return &UI{
		Output:        NewConsoleOutput(verbose),
		Input:         NewConsoleInput(),
		TableRenderer: NewConsoleTableRenderer(),
	}
}