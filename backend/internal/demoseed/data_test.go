package demoseed

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBuildPlanIsDeterministicAndComplete(t *testing.T) {
	now := time.Date(2026, time.September, 23, 15, 0, 0, 0, time.UTC)
	first := BuildPlan("demo@example.com", now)
	second := BuildPlan("demo@example.com", now)

	if !reflect.DeepEqual(first, second) {
		t.Fatal("BuildPlan returned different data for identical inputs")
	}
	if len(first.Accounts) != 2 || len(first.Categories) != 8 || len(first.Transactions) != 21 {
		t.Fatalf(
			"unexpected seed counts: accounts=%d categories=%d transactions=%d",
			len(first.Accounts),
			len(first.Categories),
			len(first.Transactions),
		)
	}
	if len(first.Imports) != 2 || len(first.Imports[1].Errors) != 2 {
		t.Fatalf("unexpected import history: %#v", first.Imports)
	}
}

func TestBuildPlanCheckingBalanceMatchesTransactions(t *testing.T) {
	plan := BuildPlan(
		"demo@example.com",
		time.Date(2026, time.September, 23, 15, 0, 0, 0, time.UTC),
	)

	var netCents int64
	for _, transaction := range plan.Transactions {
		cents := decimalCents(t, transaction.Amount)
		switch transaction.TransactionType {
		case "income":
			netCents += cents
		case "expense":
			netCents -= cents
		}
	}

	if got, want := netCents, decimalCents(t, plan.Accounts[0].CurrentBalance); got != want {
		t.Fatalf("transaction net = %d cents, checking balance = %d cents", got, want)
	}
}

func TestBuildPlanUsesLatestThreeMonthsWithoutFutureDates(t *testing.T) {
	now := time.Date(2026, time.September, 23, 15, 0, 0, 0, time.UTC)
	plan := BuildPlan("demo@example.com", now)
	months := map[string]bool{}

	for _, transaction := range plan.Transactions {
		if transaction.TransactionDate.After(now) {
			t.Fatalf("future transaction date: %s", transaction.TransactionDate)
		}
		months[transaction.TransactionDate.Format("2006-01")] = true
	}

	for _, month := range []string{"2026-09", "2026-08", "2026-07"} {
		if !months[month] {
			t.Fatalf("missing seed transactions for %s", month)
		}
	}
}

func decimalCents(t *testing.T, value string) int64 {
	t.Helper()
	parts := strings.Split(value, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		t.Fatalf("invalid fixed decimal %q", value)
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		t.Fatalf("parse whole amount %q: %v", value, err)
	}
	fraction, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		t.Fatalf("parse fractional amount %q: %v", value, err)
	}
	return whole*100 + fraction
}
