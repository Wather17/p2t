package p2t_test

import (
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
)

func TestGoalProgressAndRemaining(t *testing.T) {
	goal := p2t.Goal{
		Name:                "Reserva de emergência",
		TargetAmount:        10000,
		Deadline:            "2027-12",
		CurrentBalance:      2500,
		MonthlyContribution: 500,
	}

	if err := goal.Validate(); err != nil {
		t.Fatalf("meta deveria ser valida: %v", err)
	}
	if got := goal.ProgressPercent(); got != 25 {
		t.Errorf("progresso = %.2f, esperado 25", got)
	}
	if got := goal.RemainingAmount(); got != 7500 {
		t.Errorf("saldo restante = %.2f, esperado 7500", got)
	}
}

func TestRecurringMonthlyEquivalent(t *testing.T) {
	monthly := p2t.RecurringCommitment{
		Name:         "Streaming",
		Amount:       50,
		Period:       p2t.RecurringMonthly,
		Purpose:      p2t.PurposeLife,
		Essentiality: p2t.EssentialityDiscretionary,
		Status:       p2t.RecurringActive,
	}
	yearly := monthly
	yearly.Amount = 1200
	yearly.Period = p2t.RecurringYearly

	if err := monthly.Validate(); err != nil {
		t.Fatalf("compromisso mensal deveria ser valido: %v", err)
	}
	if err := yearly.Validate(); err != nil {
		t.Fatalf("compromisso anual deveria ser valido: %v", err)
	}
	if got := monthly.MonthlyEquivalent(); got != 50 {
		t.Errorf("equivalente mensal = %.2f, esperado 50", got)
	}
	if got := yearly.MonthlyEquivalent(); got != 100 {
		t.Errorf("equivalente anual = %.2f, esperado 100", got)
	}
}

func TestCalculateFinancialSnapshot(t *testing.T) {
	metrics, err := p2t.CalculateFinancialSnapshot(5000, 1800, 1500, 200)
	if err != nil {
		t.Fatalf("calculo do fechamento falhou: %v", err)
	}
	if metrics.RecurringCommitmentTotal != 1800 {
		t.Errorf("compromissos = %.2f, esperado 1800", metrics.RecurringCommitmentTotal)
	}
	if metrics.GoalCapacity != 3200 {
		t.Errorf("capacidade de metas = %.2f, esperado 3200", metrics.GoalCapacity)
	}
	if metrics.FreeMargin != 1500 {
		t.Errorf("margem livre = %.2f, esperado 1500", metrics.FreeMargin)
	}
}

func TestValidateReferenceMonth(t *testing.T) {
	valid := []string{"2026-01", "1999-12"}
	for _, value := range valid {
		if err := p2t.ValidateReferenceMonth(value); err != nil {
			t.Errorf("competencia %q deveria ser valida: %v", value, err)
		}
	}

	invalid := []string{"", "2026-1", "2026-13", "26-01"}
	for _, value := range invalid {
		if err := p2t.ValidateReferenceMonth(value); err == nil {
			t.Errorf("competencia %q deveria ser invalida", value)
		}
	}
}
