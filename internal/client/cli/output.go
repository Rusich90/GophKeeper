package cli

import (
	"fmt"
	"log/slog"
	"os"
)

// ColorOutput обеспечивает цветной вывод в консоль
type ColorOutput struct {
	logger *slog.Logger
}

// NewColorOutput создает новый экземпляр ColorOutput
func NewColorOutput(verbose bool) *ColorOutput {
	level := slog.LevelError
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &ColorOutput{
		logger: logger,
	}
}

// Success выводит успешное сообщение зеленым цветом
func (c *ColorOutput) Success(message string) {
	fmt.Printf("\033[32m✓ %s\033[0m\n", message)
}

// Successf выводит успешное сообщение с форматированием
func (c *ColorOutput) Successf(format string, args ...interface{}) {
	fmt.Printf("\033[32m✓ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Error выводит сообщение об ошибке красным цветом
func (c *ColorOutput) Error(message string) {
	fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", message)
}

// Errorf выводит сообщение об ошибке с форматированием
func (c *ColorOutput) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m✗ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Info выводит информационное сообщение синим цветом
func (c *ColorOutput) Info(message string) {
	fmt.Printf("\033[34mℹ %s\033[0m\n", message)
}

// Infof выводит информационное сообщение с форматированием
func (c *ColorOutput) Infof(format string, args ...interface{}) {
	fmt.Printf("\033[34mℹ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Warning выводит предупреждение желтым цветом
func (c *ColorOutput) Warning(message string) {
	fmt.Printf("\033[33m⚠ %s\033[0m\n", message)
}

// Warningf выводит предупреждение с форматированием
func (c *ColorOutput) Warningf(format string, args ...interface{}) {
	fmt.Printf("\033[33m⚠ %s\033[0m\n", fmt.Sprintf(format, args...))
}

// Plain выводит обычное сообщение без цвета
func (c *ColorOutput) Plain(message string) {
	fmt.Println(message)
}

// Plainf выводит обычное сообщение с форматированием
func (c *ColorOutput) Plainf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Debug выводит отладочное сообщение (только в verbose режиме)
func (c *ColorOutput) Debug(message string) {
	c.logger.Debug(message)
}

// Debugf выводит отладочное сообщение с форматированием
func (c *ColorOutput) Debugf(format string, args ...interface{}) {
	c.logger.Debug(fmt.Sprintf(format, args...))
}

// GetLogger возвращает логгер
func (c *ColorOutput) GetLogger() *slog.Logger {
	return c.logger
}
