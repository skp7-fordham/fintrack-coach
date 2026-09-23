package demoseed

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type AccountSeed struct {
	ID             string
	Name           string
	AccountType    string
	Currency       string
	CurrentBalance string
}

type CategorySeed struct {
	ID           string
	Name         string
	CategoryType string
	Color        string
	Icon         string
}

type TransactionSeed struct {
	ID                string
	AccountID         string
	CategoryID        string
	Description       string
	Merchant          string
	Amount            string
	TransactionType   string
	TransactionStatus string
	TransactionDate   time.Time
	Notes             string
}

type ImportErrorSeed struct {
	RowNumber    int
	ErrorMessage string
	RawRow       map[string]string
}

type ImportSeed struct {
	ID               string
	AccountID        string
	OriginalFilename string
	Status           string
	TotalRows        int
	SuccessfulRows   int
	FailedRows       int
	CreatedAt        time.Time
	Errors           []ImportErrorSeed
}

type Plan struct {
	Accounts     []AccountSeed
	Categories   []CategorySeed
	Transactions []TransactionSeed
	Imports      []ImportSeed
}

func BuildPlan(email string, now time.Time) Plan {
	email = strings.ToLower(strings.TrimSpace(email))
	now = now.UTC()

	checkingID := stableUUID(email, "account:checking")
	savingsID := stableUUID(email, "account:savings")

	categoryIDs := map[string]string{}
	category := func(name, categoryType, color, icon string) CategorySeed {
		id := stableUUID(email, "category:"+strings.ToLower(name))
		categoryIDs[name] = id
		return CategorySeed{
			ID:           id,
			Name:         name,
			CategoryType: categoryType,
			Color:        color,
			Icon:         icon,
		}
	}

	categories := []CategorySeed{
		category("Groceries", "expense", "#0f766e", "G"),
		category("Dining", "expense", "#ea580c", "D"),
		category("Rent", "expense", "#7c3aed", "R"),
		category("Transport", "expense", "#0369a1", "T"),
		category("Shopping", "expense", "#be185d", "S"),
		category("Utilities", "expense", "#475569", "U"),
		category("Salary", "income", "#047857", "$"),
		category("Freelance", "income", "#0d9488", "F"),
	}

	type transactionInput struct {
		key, description, merchant, amount, transactionType, category string
		monthOffset, day                                              int
	}
	inputs := []transactionInput{
		{"current-salary", "Monthly salary", "Northstar Studio", "4200.00", "income", "Salary", 0, 1},
		{"current-rent", "Apartment rent", "Riverside Properties", "1450.00", "expense", "Rent", 0, 3},
		{"current-groceries", "Weekly groceries", "Green Market", "126.42", "expense", "Groceries", 0, 7},
		{"current-dining", "Dinner with friends", "Harbor Kitchen", "58.75", "expense", "Dining", 0, 12},
		{"current-transport", "Transit pass", "Metro Transit", "42.20", "expense", "Transport", 0, 18},
		{"current-utilities", "Electric and internet", "City Utilities", "118.60", "expense", "Utilities", 0, 22},

		{"previous-salary", "Monthly salary", "Northstar Studio", "4200.00", "income", "Salary", -1, 1},
		{"previous-freelance", "Website project", "Maple & Co.", "650.00", "income", "Freelance", -1, 6},
		{"previous-rent", "Apartment rent", "Riverside Properties", "1450.00", "expense", "Rent", -1, 3},
		{"previous-groceries", "Groceries", "Green Market", "310.80", "expense", "Groceries", -1, 9},
		{"previous-dining", "Dining out", "Various restaurants", "142.25", "expense", "Dining", -1, 14},
		{"previous-transport", "Transit and rides", "Metro Transit", "96.40", "expense", "Transport", -1, 18},
		{"previous-shopping", "Home essentials", "Market Square", "185.99", "expense", "Shopping", -1, 22},
		{"previous-utilities", "Electric and internet", "City Utilities", "121.30", "expense", "Utilities", -1, 25},

		{"two-ago-salary", "Monthly salary", "Northstar Studio", "4200.00", "income", "Salary", -2, 1},
		{"two-ago-rent", "Apartment rent", "Riverside Properties", "1450.00", "expense", "Rent", -2, 3},
		{"two-ago-groceries", "Groceries", "Green Market", "287.45", "expense", "Groceries", -2, 8},
		{"two-ago-dining", "Dining out", "Various restaurants", "119.70", "expense", "Dining", -2, 13},
		{"two-ago-transport", "Transit and rides", "Metro Transit", "88.20", "expense", "Transport", -2, 17},
		{"two-ago-shopping", "Household purchase", "Market Square", "79.99", "expense", "Shopping", -2, 21},
		{"two-ago-utilities", "Electric and internet", "City Utilities", "116.80", "expense", "Utilities", -2, 26},
	}

	transactions := make([]TransactionSeed, 0, len(inputs))
	for _, input := range inputs {
		transactions = append(transactions, TransactionSeed{
			ID:                stableUUID(email, "transaction:"+input.key),
			AccountID:         checkingID,
			CategoryID:        categoryIDs[input.category],
			Description:       input.description,
			Merchant:          input.merchant,
			Amount:            input.amount,
			TransactionType:   input.transactionType,
			TransactionStatus: "completed",
			TransactionDate:   relativeMonthDate(now, input.monthOffset, input.day),
			Notes:             "Demo sample transaction",
		})
	}

	return Plan{
		Accounts: []AccountSeed{
			{
				ID:             checkingID,
				Name:           "Everyday Checking",
				AccountType:    "checking",
				Currency:       "USD",
				CurrentBalance: "7005.15",
			},
			{
				ID:             savingsID,
				Name:           "Emergency Savings",
				AccountType:    "savings",
				Currency:       "USD",
				CurrentBalance: "12500.00",
			},
		},
		Categories:   categories,
		Transactions: transactions,
		Imports: []ImportSeed{
			{
				ID:               stableUUID(email, "import:completed"),
				AccountID:        checkingID,
				OriginalFilename: "demo-transactions.csv",
				Status:           "completed",
				TotalRows:        12,
				SuccessfulRows:   12,
				CreatedAt:        now.AddDate(0, 0, -4),
			},
			{
				ID:               stableUUID(email, "import:mixed"),
				AccountID:        checkingID,
				OriginalFilename: "demo-mixed-import.csv",
				Status:           "completed_with_errors",
				TotalRows:        10,
				SuccessfulRows:   8,
				FailedRows:       2,
				CreatedAt:        now.AddDate(0, 0, -12),
				Errors: []ImportErrorSeed{
					{
						RowNumber:    4,
						ErrorMessage: "category does not exist",
						RawRow: map[string]string{
							"date": relativeMonthDate(now, -1, 12).Format("2006-01-02"), "description": "Gym membership",
							"amount": "49.00", "type": "expense", "category": "Fitness",
						},
					},
					{
						RowNumber:    9,
						ErrorMessage: "amount must be a positive decimal with at most 2 places",
						RawRow: map[string]string{
							"date": relativeMonthDate(now, -1, 19).Format("2006-01-02"), "description": "Coffee",
							"amount": "four dollars", "type": "expense", "category": "Dining",
						},
					},
				},
			},
		},
	}
}

func relativeMonthDate(now time.Time, monthOffset, day int) time.Time {
	first := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, time.UTC).AddDate(0, monthOffset, 0)
	if monthOffset == 0 && day > now.Day() {
		day = now.Day()
	}
	return first.AddDate(0, 0, day-1)
}

func stableUUID(email, key string) string {
	sum := sha256.Sum256([]byte("fintrack-demo:" + email + ":" + key))
	bytes := sum[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	)
}
