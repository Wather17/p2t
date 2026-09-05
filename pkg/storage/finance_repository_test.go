package storage_test

import (
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
)

func TestRepository_FinanceEntitiesAndSnapshot(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco: %v", err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)

	goalID, err := repo.CreateGoal(p2t.Goal{
		Name: "Reserva", TargetAmount: 10000, Deadline: "2027-12",
		CurrentBalance: 2500, MonthlyContribution: 500,
	})
	if err != nil {
		t.Fatalf("falha ao criar meta: %v", err)
	}
	if goalID <= 0 {
		t.Fatalf("id da meta deveria ser positivo: %d", goalID)
	}

	recurringID, err := repo.CreateRecurring(p2t.RecurringCommitment{
		Name: "Seguro anual", Amount: 1200, Period: p2t.RecurringYearly,
		Purpose: p2t.PurposeLife, Essentiality: p2t.EssentialityEssential,
		BillingDay: 5, Status: p2t.RecurringActive,
	})
	if err != nil {
		t.Fatalf("falha ao criar compromisso: %v", err)
	}
	if recurringID <= 0 {
		t.Fatalf("id do compromisso deveria ser positivo: %d", recurringID)
	}

	items, err := repo.GetActiveRecurring()
	if err != nil {
		t.Fatalf("falha ao listar compromissos ativos: %v", err)
	}
	if len(items) != 1 || items[0].MonthlyEquivalent() != 100 {
		t.Fatalf("compromisso anual deveria equivaler a R$ 100/mês: %+v", items)
	}

	metrics, err := p2t.CalculateFinancialSnapshot(5000, 100, 500, 0)
	if err != nil {
		t.Fatalf("falha ao calcular métricas: %v", err)
	}
	snapshot := p2t.FinancialSnapshot{
		ReferenceMonth: "2026-08", ReceivedIncome: 5000, ReserveBalance: 2500,
		PlannedGoalContributions: 500, ExceptionalCosts: 0,
		RecurringCommitmentTotal: metrics.RecurringCommitmentTotal,
		GoalCapacity:             metrics.GoalCapacity, FreeMargin: metrics.FreeMargin,
	}
	if _, err := repo.SaveFinancialSnapshot(snapshot, []storage.GoalSnapshotProgressRecord{{
		GoalID: goalID, CurrentBalance: 2500, PlannedContribution: 500,
	}}); err != nil {
		t.Fatalf("falha ao salvar fechamento: %v", err)
	}

	loaded, err := repo.GetFinancialSnapshot("2026-08")
	if err != nil {
		t.Fatalf("falha ao buscar fechamento: %v", err)
	}
	if loaded == nil || loaded.FreeMargin != 4400 {
		t.Fatalf("fechamento inesperado: %+v", loaded)
	}
	if len(loaded.Goals) != 1 || loaded.Goals[0].GoalID != goalID {
		t.Fatalf("progresso de meta inesperado: %+v", loaded.Goals)
	}

	// Salvar a mesma competência deve atualizar, não duplicar.
	snapshot.ReceivedIncome = 5200
	snapshot.GoalCapacity = 5100
	snapshot.FreeMargin = 4600
	if _, err := repo.SaveFinancialSnapshot(snapshot, nil); err != nil {
		t.Fatalf("falha ao atualizar fechamento: %v", err)
	}
	updated, err := repo.GetFinancialSnapshot("2026-08")
	if err != nil {
		t.Fatalf("falha ao reler fechamento atualizado: %v", err)
	}
	if updated.ReceivedIncome != 5200 || len(updated.Goals) != 0 {
		t.Fatalf("fechamento não foi substituído corretamente: %+v", updated)
	}
}
