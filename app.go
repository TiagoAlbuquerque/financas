package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx   context.Context
	store *Store
}

// NewApp creates a new App application struct
func NewApp() *App {
	store := NewStore()
	// Seed initial data if database is empty so the user starts with a populated dashboard
	_ = store.SeedMockData()
	return &App{
		store: store,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetOverallSummary returns the global totals and yearly totals
func (a *App) GetOverallSummary() (*OverallSummary, error) {
	return a.store.GetOverallSummary()
}

// GetYearSummary returns totals and monthly breakdown for a specific year
func (a *App) GetYearSummary(year int) (*YearSummary, error) {
	return a.store.GetYearSummary(year)
}

// GetMonthSummary returns category breakdowns and totals for a month
func (a *App) GetMonthSummary(year int, month int) (*MonthSummary, error) {
	return a.store.GetMonthSummary(year, month)
}

// GetMonthTransactions returns all transactions for a specific month
func (a *App) GetMonthTransactions(year int, month int) (*MonthTransactions, error) {
	return a.store.GetMonthTransactions(year, month)
}

// SaveTransaction adds or updates a transaction and updates all affected summaries
func (a *App) SaveTransaction(t Transaction) (Transaction, error) {
	return a.store.SaveTransaction(t)
}

// DeleteTransaction removes a transaction and updates all summaries
func (a *App) DeleteTransaction(id string, year int, month int) error {
	return a.store.DeleteTransaction(id, year, month)
}

// GetDataDir returns the directory path where JSON files are saved
func (a *App) GetDataDir() string {
	return a.store.DataDir
}

// Greet returns a greeting for the given name (kept for compatibility)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Olá %s! O sistema de finanças está rodando localmente em: %s", name, a.store.DataDir)
}

// CheckForUpdate calls the internal updater to check if a new version is available on GitHub
func (a *App) CheckForUpdate() (*UpdateInfo, error) {
	return CheckForUpdate()
}

// ApplyUpdate downloads the new release asset and replaces the active binary
func (a *App) ApplyUpdate(info *UpdateInfo) error {
	if info == nil {
		return fmt.Errorf("informações de atualização inválidas")
	}
	return DownloadAndInstallUpdate(info)
}

// RestartApp restarts the application by spawning the new binary and exiting
func (a *App) RestartApp() error {
	return TriggerRestart()
}

