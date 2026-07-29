package coach

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/aiprovider"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t.UTC() }
}

func TestCurrentMonthResolutionUsesInjectedClock(t *testing.T) {
	now := time.Date(2026, time.July, 29, 15, 4, 5, 0, time.UTC)
	if got := CurrentMonth(now); got != "2026-07" {
		t.Fatalf("CurrentMonth = %q, want 2026-07", got)
	}

	prompt := buildSystemPrompt(now)
	if !strings.Contains(prompt, "Current date: 2026-07-29") {
		t.Fatalf("system prompt missing current date:\n%s", prompt)
	}
	if !strings.Contains(prompt, "this month") || !strings.Contains(prompt, "2026-07") {
		t.Fatalf("system prompt missing this-month grounding:\n%s", prompt)
	}
}

func TestPreviousMonthResolutionIncludingJanuary(t *testing.T) {
	july := time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC)
	if got := PreviousMonth(july); got != "2026-06" {
		t.Fatalf("PreviousMonth(July) = %q, want 2026-06", got)
	}

	january := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)
	if got := PreviousMonth(january); got != "2025-12" {
		t.Fatalf("PreviousMonth(January) = %q, want 2025-12", got)
	}

	prompt := buildSystemPrompt(january)
	if !strings.Contains(prompt, "previous month") || !strings.Contains(prompt, "2025-12") {
		t.Fatalf("system prompt missing previous-month grounding:\n%s", prompt)
	}
}

func TestDateGroundingDoesNotInventArbitraryYear(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	current := CurrentMonth(now)
	previous := PreviousMonth(now)
	prompt := buildSystemPrompt(now)

	if current != "2026-07" || previous != "2026-06" {
		t.Fatalf("unexpected months current=%s previous=%s", current, previous)
	}
	if strings.Contains(current, "2023") || strings.Contains(previous, "2023") || strings.Contains(prompt, "2023") {
		t.Fatalf("arbitrary year 2023 present in date grounding output")
	}

	llm := &stubLLM{responses: []aiprovider.GenerateResponse{{Content: "Your current total balance is available from your accounts."}}}
	agent := NewAgentWithClock(llm, &recordingTools{}, "test-model", 5, fixedClock(now), testLogger())
	if _, err := agent.Run(context.Background(), "11111111-1111-1111-1111-111111111111", nil, "What is my current total balance?"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(llm.calls) == 0 {
		t.Fatal("expected llm call")
	}
	system := llm.calls[0].Messages[0].Content
	if !strings.Contains(system, "Current date: 2026-07-29") {
		t.Fatalf("agent system prompt missing current date: %s", system)
	}
	if strings.Contains(system, "2023") {
		t.Fatalf("agent system prompt invented 2023: %s", system)
	}
}

func TestBalanceSnapshotFollowUpDoesNotClaimHistoricalBalances(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	prompt := buildSystemPrompt(now)

	required := []string{
		"current_balance is a present snapshot",
		"cannot be historically compared",
		"Do not misuse dashboard summaries to claim historical account balances",
		"cannot accurately compare historical total balances",
		"income, expenses, and net cash flow",
		"2026-06",
		"2026-07",
	}
	for _, needle := range required {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("prompt missing %q:\n%s", needle, prompt)
		}
	}

	llm := &stubLLM{
		responses: []aiprovider.GenerateResponse{
			{Content: "I cannot accurately compare historical total balances from current account snapshots, but I can compare income, expenses, and net cash flow for June and July 2026."},
		},
	}
	history := []domain.CoachMessage{
		{Role: domain.CoachMessageRoleUser, Content: "What is my current total balance?"},
		{Role: domain.CoachMessageRoleAssistant, Content: "Based on your recorded accounts, your current total balance is $1,250.00."},
	}
	agent := NewAgentWithClock(llm, &recordingTools{}, "test-model", 5, fixedClock(now), testLogger())
	result, err := agent.Run(
		context.Background(),
		"11111111-1111-1111-1111-111111111111",
		history,
		"How does that compare with the previous month?",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(strings.ToLower(result.Content), "september 2023") ||
		strings.Contains(strings.ToLower(result.Content), "october 2023") {
		t.Fatalf("response claimed arbitrary historical months: %s", result.Content)
	}
	system := llm.calls[0].Messages[0].Content
	if !strings.Contains(system, "cannot accurately compare historical total balances") {
		t.Fatal("system prompt did not include balance follow-up guidance")
	}
	if len(llm.calls[0].Messages) < 4 {
		t.Fatalf("expected history to be included, got %d messages", len(llm.calls[0].Messages))
	}
}

func TestJuly2026SpendingFollowUpComparesWithJune2026(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	if CurrentMonth(now) != "2026-07" {
		t.Fatalf("current month = %s", CurrentMonth(now))
	}
	if PreviousMonth(now) != "2026-06" {
		t.Fatalf("previous month = %s", PreviousMonth(now))
	}

	prompt := buildSystemPrompt(now)
	if !strings.Contains(prompt, "(2026-07)") || !strings.Contains(prompt, "(2026-06)") {
		t.Fatalf("prompt should ground this/previous month to July/June 2026:\n%s", prompt)
	}

	registry := NewToolRegistry(nil, nil, nil, testLogger())
	var summaryDesc, trendsDesc, accountsDesc string
	for _, def := range registry.Definitions() {
		switch def.Function.Name {
		case toolGetDashboardSummary:
			summaryDesc = def.Function.Description
		case toolGetMonthlyTrends:
			trendsDesc = def.Function.Description
		case toolListAccounts:
			accountsDesc = def.Function.Description
		}
	}
	if !strings.Contains(summaryDesc, "current total balance snapshot") ||
		!strings.Contains(summaryDesc, "not a historical month-end balance") {
		t.Fatalf("dashboard summary tool description unclear: %s", summaryDesc)
	}
	if !strings.Contains(trendsDesc, "income, expense, and net cash flow") ||
		!strings.Contains(trendsDesc, "does not provide historical account balances") {
		t.Fatalf("monthly trends tool description unclear: %s", trendsDesc)
	}
	if !strings.Contains(accountsDesc, "current balances only") {
		t.Fatalf("list_accounts tool description unclear: %s", accountsDesc)
	}
}
