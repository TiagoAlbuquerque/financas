package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Transaction represents a single income or expense record
type Transaction struct {
	ID          string    `json:"id"`
	Date        string    `json:"date"` // YYYY-MM-DD
	Type        string    `json:"type"` // "earning" or "expense"
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
}

// MonthTransactions holds detailed transactions for a specific month
type MonthTransactions struct {
	Year         int           `json:"year"`
	Month        int           `json:"month"`
	Transactions []Transaction `json:"transactions"`
}

// MonthSummary holds summarized statistics for a month
type MonthSummary struct {
	Year               int                `json:"year"`
	Month              int                `json:"month"`
	Earnings           float64            `json:"earnings"`
	Expenses           float64            `json:"expenses"`
	EarningsByCategory map[string]float64 `json:"earnings_by_category"`
	ExpensesByCategory map[string]float64 `json:"expenses_by_category"`
}

// MonthTotal holds simple totals for a month in the yearly summary
type MonthTotal struct {
	Earnings float64 `json:"earnings"`
	Expenses float64 `json:"expenses"`
}

// YearSummary holds summarized statistics for a year
type YearSummary struct {
	Year               int                   `json:"year"`
	Earnings           float64               `json:"earnings"`
	Expenses           float64               `json:"expenses"`
	Months             map[string]MonthTotal `json:"months"` // "01", "02", etc.
	EarningsByCategory map[string]float64    `json:"earnings_by_category"`
	ExpensesByCategory map[string]float64    `json:"expenses_by_category"`
}

// YearTotal holds simple totals for a year in the overall summary
type YearTotal struct {
	Earnings float64 `json:"earnings"`
	Expenses float64 `json:"expenses"`
}

// OverallSummary holds overall statistics across all years
type OverallSummary struct {
	Earnings float64              `json:"earnings"`
	Expenses float64              `json:"expenses"`
	Years    map[string]YearTotal `json:"years"` // "2026", etc.
}

// Store handles the JSON file database
type Store struct {
	DataDir string
}

// NewStore initializes a store
func NewStore() *Store {
	return &Store{
		DataDir: getDataDir(),
	}
}

// getDataDir returns the data folder path, trying same-dir portable first, fallback to user config dir.
func getDataDir() string {
	// Try portable mode (directory containing the executable)
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		// Don't write to /tmp or system directories if building there
		if !strings.Contains(exeDir, "/tmp") && !strings.Contains(exeDir, "Antigravity") {
			localDataDir := filepath.Join(exeDir, "data")
			if err := os.MkdirAll(localDataDir, 0755); err == nil {
				// Test write permissions
				testFile := filepath.Join(localDataDir, ".test")
				if err := os.WriteFile(testFile, []byte(""), 0644); err == nil {
					os.Remove(testFile)
					return localDataDir
				}
			}
		}
	}

	// Fallback to standard UserConfigDir
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	appDir := filepath.Join(configDir, "FinancasPersonalApp", "data")
	_ = os.MkdirAll(appDir, 0755)
	return appDir
}

// generateID generates a pseudo-UUID
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Helper to parse date string YYYY-MM-DD
func parseDate(dateStr string) (int, int, error) {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("formato de data invalido: %s", dateStr)
	}
	y, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return y, m, nil
}

// GetOverallSummary returns the main summary
func (s *Store) GetOverallSummary() (*OverallSummary, error) {
	path := filepath.Join(s.DataDir, "summary.json")
	var summary OverallSummary
	if err := readJSON(path, &summary); err != nil {
		// Return empty summary if file doesn't exist
		return &OverallSummary{
			Earnings: 0,
			Expenses: 0,
			Years:    make(map[string]YearTotal),
		}, nil
	}
	if summary.Years == nil {
		summary.Years = make(map[string]YearTotal)
	}
	return &summary, nil
}

// GetYearSummary returns the summary of a specific year
func (s *Store) GetYearSummary(year int) (*YearSummary, error) {
	path := filepath.Join(s.DataDir, strconv.Itoa(year), "summary.json")
	var summary YearSummary
	if err := readJSON(path, &summary); err != nil {
		return &YearSummary{
			Year:     year,
			Earnings: 0,
			Expenses: 0,
			Months:   make(map[string]MonthTotal),
		}, nil
	}
	if summary.Months == nil {
		summary.Months = make(map[string]MonthTotal)
	}
	return &summary, nil
}

// GetMonthSummary returns the summary of a specific month
func (s *Store) GetMonthSummary(year int, month int) (*MonthSummary, error) {
	monthStr := fmt.Sprintf("%02d", month)
	path := filepath.Join(s.DataDir, strconv.Itoa(year), monthStr, "summary.json")
	var summary MonthSummary
	if err := readJSON(path, &summary); err != nil {
		return &MonthSummary{
			Year:               year,
			Month:              month,
			Earnings:           0,
			Expenses:           0,
			EarningsByCategory: make(map[string]float64),
			ExpensesByCategory: make(map[string]float64),
		}, nil
	}
	if summary.EarningsByCategory == nil {
		summary.EarningsByCategory = make(map[string]float64)
	}
	if summary.ExpensesByCategory == nil {
		summary.ExpensesByCategory = make(map[string]float64)
	}
	return &summary, nil
}

// GetMonthTransactions returns all transactions for a month
func (s *Store) GetMonthTransactions(year int, month int) (*MonthTransactions, error) {
	monthStr := fmt.Sprintf("%02d", month)
	path := filepath.Join(s.DataDir, strconv.Itoa(year), monthStr, "details.json")
	var details MonthTransactions
	if err := readJSON(path, &details); err != nil {
		return &MonthTransactions{
			Year:         year,
			Month:        month,
			Transactions: []Transaction{},
		}, nil
	}
	if details.Transactions == nil {
		details.Transactions = []Transaction{}
	}
	return &details, nil
}

// SaveTransaction adds or updates a transaction and updates all affected summaries
func (s *Store) SaveTransaction(t Transaction) (Transaction, error) {
	y, m, err := parseDate(t.Date)
	if err != nil {
		return t, err
	}

	if t.ID == "" {
		t.ID = generateID()
	}

	monthStr := fmt.Sprintf("%02d", m)
	dirPath := filepath.Join(s.DataDir, strconv.Itoa(y), monthStr)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return t, err
	}

	// Read existing transactions
	detailsPath := filepath.Join(dirPath, "details.json")
	var details MonthTransactions
	_ = readJSON(detailsPath, &details)
	details.Year = y
	details.Month = m

	// Check if update or insert
	found := false
	for idx, item := range details.Transactions {
		if item.ID == t.ID {
			details.Transactions[idx] = t
			found = true
			break
		}
	}
	if !found {
		details.Transactions = append(details.Transactions, t)
	}

	// Save transactions
	if err := writeJSON(detailsPath, &details); err != nil {
		return t, err
	}

	// Recompute summaries
	err = s.recomputeAll(y, m)
	return t, err
}

// DeleteTransaction deletes a transaction and updates summaries
func (s *Store) DeleteTransaction(id string, year int, month int) error {
	monthStr := fmt.Sprintf("%02d", month)
	detailsPath := filepath.Join(s.DataDir, strconv.Itoa(year), monthStr, "details.json")

	var details MonthTransactions
	if err := readJSON(detailsPath, &details); err != nil {
		return fmt.Errorf("transações não encontradas para este mês")
	}

	newTransactions := []Transaction{}
	found := false
	for _, item := range details.Transactions {
		if item.ID == id {
			found = true
			continue
		}
		newTransactions = append(newTransactions, item)
	}

	if !found {
		return fmt.Errorf("transação não encontrada")
	}

	details.Transactions = newTransactions
	if err := writeJSON(detailsPath, &details); err != nil {
		return err
	}

	return s.recomputeAll(year, month)
}

// recomputeAll updates the monthly, yearly, and overall summaries based on details files.
func (s *Store) recomputeAll(year int, month int) error {
	monthStr := fmt.Sprintf("%02d", month)

	// 1. Recompute Monthly Summary
	detailsPath := filepath.Join(s.DataDir, strconv.Itoa(year), monthStr, "details.json")
	var details MonthTransactions
	_ = readJSON(detailsPath, &details)

	monthSummary := MonthSummary{
		Year:               year,
		Month:              month,
		EarningsByCategory: make(map[string]float64),
		ExpensesByCategory: make(map[string]float64),
	}

	for _, t := range details.Transactions {
		if t.Type == "earning" {
			monthSummary.Earnings += t.Amount
			monthSummary.EarningsByCategory[t.Category] += t.Amount
		} else {
			monthSummary.Expenses += t.Amount
			monthSummary.ExpensesByCategory[t.Category] += t.Amount
		}
	}

	monthSummaryPath := filepath.Join(s.DataDir, strconv.Itoa(year), monthStr, "summary.json")
	if err := writeJSON(monthSummaryPath, &monthSummary); err != nil {
		return err
	}

	// 2. Recompute Yearly Summary
	yearDir := filepath.Join(s.DataDir, strconv.Itoa(year))
	yearSummary := YearSummary{
		Year:               year,
		Months:             make(map[string]MonthTotal),
		EarningsByCategory: make(map[string]float64),
		ExpensesByCategory: make(map[string]float64),
	}

	// Scan month directories (01-12)
	for mIdx := 1; mIdx <= 12; mIdx++ {
		mStr := fmt.Sprintf("%02d", mIdx)
		mSumPath := filepath.Join(yearDir, mStr, "summary.json")
		var mSum MonthSummary
		if err := readJSON(mSumPath, &mSum); err == nil {
			yearSummary.Earnings += mSum.Earnings
			yearSummary.Expenses += mSum.Expenses
			yearSummary.Months[mStr] = MonthTotal{
				Earnings: mSum.Earnings,
				Expenses: mSum.Expenses,
			}
			for cat, amt := range mSum.EarningsByCategory {
				yearSummary.EarningsByCategory[cat] += amt
			}
			for cat, amt := range mSum.ExpensesByCategory {
				yearSummary.ExpensesByCategory[cat] += amt
			}
		}
	}

	yearSummaryPath := filepath.Join(yearDir, "summary.json")
	if err := writeJSON(yearSummaryPath, &yearSummary); err != nil {
		return err
	}

	// 3. Recompute Overall Summary
	overallSummary := OverallSummary{
		Years: make(map[string]YearTotal),
	}

	// Search for all 4-digit folders in DataDir
	files, err := os.ReadDir(s.DataDir)
	if err == nil {
		yearRegexp := regexp.MustCompile(`^\d{4}$`)
		for _, file := range files {
			if file.IsDir() && yearRegexp.MatchString(file.Name()) {
				yPath := filepath.Join(s.DataDir, file.Name(), "summary.json")
				var ySum YearSummary
				if err := readJSON(yPath, &ySum); err == nil {
					overallSummary.Earnings += ySum.Earnings
					overallSummary.Expenses += ySum.Expenses
					overallSummary.Years[file.Name()] = YearTotal{
						Earnings: ySum.Earnings,
						Expenses: ySum.Expenses,
					}
				}
			}
		}
	}

	overallSummaryPath := filepath.Join(s.DataDir, "summary.json")
	return writeJSON(overallSummaryPath, &overallSummary)
}

// JSON helpers
func readJSON(path string, target interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(target)
}

func writeJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Helper method to add some mock data if the database is completely empty
func (s *Store) SeedMockData() error {
	summary, err := s.GetOverallSummary()
	if err == nil && len(summary.Years) > 0 {
		// Already has data, no need to seed
		return nil
	}

	// Let's seed for current and past year
	now := time.Now()
	thisYear := now.Year()
	prevYear := thisYear - 1

	mockTransactions := []Transaction{
		// Previous year
		{Date: fmt.Sprintf("%d-01-10", prevYear), Type: "earning", Category: "Salário", Description: "Salário Janeiro", Amount: 5000.0},
		{Date: fmt.Sprintf("%d-01-12", prevYear), Type: "expense", Category: "Moradia", Description: "Aluguel", Amount: 1500.0},
		{Date: fmt.Sprintf("%d-01-15", prevYear), Type: "expense", Category: "Alimentação", Description: "Supermercado", Amount: 600.0},
		{Date: fmt.Sprintf("%d-02-10", prevYear), Type: "earning", Category: "Salário", Description: "Salário Fevereiro", Amount: 5000.0},
		{Date: fmt.Sprintf("%d-02-14", prevYear), Type: "expense", Category: "Transporte", Description: "Combustível", Amount: 300.0},
		{Date: fmt.Sprintf("%d-02-20", prevYear), Type: "expense", Category: "Lazer", Description: "Cinema e Jantar", Amount: 400.0},
		
		// Current year
		{Date: fmt.Sprintf("%d-01-10", thisYear), Type: "earning", Category: "Salário", Description: "Salário Janeiro", Amount: 5500.0},
		{Date: fmt.Sprintf("%d-01-11", thisYear), Type: "expense", Category: "Moradia", Description: "Aluguel", Amount: 1600.0},
		{Date: fmt.Sprintf("%d-01-20", thisYear), Type: "expense", Category: "Alimentação", Description: "Feira Semanal", Amount: 450.0},
		{Date: fmt.Sprintf("%d-02-10", thisYear), Type: "earning", Category: "Salário", Description: "Salário Fevereiro", Amount: 5500.0},
		{Date: fmt.Sprintf("%d-02-12", thisYear), Type: "expense", Category: "Moradia", Description: "Aluguel", Amount: 1600.0},
		{Date: fmt.Sprintf("%d-02-28", thisYear), Type: "expense", Category: "Saúde", Description: "Farmácia", Amount: 180.0},
		{Date: fmt.Sprintf("%d-03-10", thisYear), Type: "earning", Category: "Salário", Description: "Salário Março", Amount: 5500.0},
		{Date: fmt.Sprintf("%d-03-15", thisYear), Type: "earning", Category: "Investimentos", Description: "Dividendos FIIs", Amount: 450.0},
		{Date: fmt.Sprintf("%d-03-20", thisYear), Type: "expense", Category: "Lazer", Description: "Viagem Fim de Semana", Amount: 1200.0},
	}

	for _, t := range mockTransactions {
		if _, err := s.SaveTransaction(t); err != nil {
			return err
		}
	}

	return nil
}
