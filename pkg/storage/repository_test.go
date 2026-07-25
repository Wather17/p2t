package storage_test

import (
	"math"
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
)

const floatTolerance = 0.0001

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < floatTolerance
}

func TestRepository_Telemetry(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ErrorDeductions: 200.0,
		InvisibleCosts:  300.0,
		ContractHours:   160.0,
		CommuteHours:    40.0,
	}
	res, err := p2t.CalculateTelemetry(input)
	if err != nil {
		t.Fatalf("falha ao calcular telemetria: %v", err)
	}

	id, err := repo.SaveTelemetry(input, res)
	if err != nil {
		t.Fatalf("falha ao salvar telemetria no banco: %v", err)
	}
	if id <= 0 {
		t.Errorf("esperado ID > 0, obtido: %d", id)
	}

	// Adiciona mais dois ciclos para testar historico de IDT
	_, _ = repo.SaveTelemetry(input, res)
	input2 := input
	input2.ErrorDeductions = 500.0 // altera IDT
	res2, _ := p2t.CalculateTelemetry(input2)
	_, _ = repo.SaveTelemetry(input2, res2)

	history, err := repo.GetRecentIDTHistory(3)
	if err != nil {
		t.Fatalf("falha ao consultar historico de IDT: %v", err)
	}

	if len(history) != 3 {
		t.Fatalf("esperado 3 registros no historico, obtido: %d", len(history))
	}

	if !almostEqual(history[2], res2.IDT) {
		t.Errorf("esperado ultimo IDT=%.2f, obtido=%.2f", res2.IDT, history[2])
	}
}

func TestRepository_Buffer(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	id, err := repo.SaveBufferCycle(300.0, 180.0, 120.0)
	if err != nil {
		t.Fatalf("falha ao salvar ciclo de buffer: %v", err)
	}
	if id <= 0 {
		t.Errorf("esperado ID > 0, obtido: %d", id)
	}

	_, _ = repo.SaveBufferCycle(300.0, 150.0, 150.0)
	_, _ = repo.SaveBufferCycle(300.0, 200.0, 100.0)

	replenishments, err := repo.GetRecentReplenishments(3)
	if err != nil {
		t.Fatalf("falha ao consultar reposicoes: %v", err)
	}

	if len(replenishments) != 3 {
		t.Fatalf("esperado 3 registros de reposicao, obtido: %d", len(replenishments))
	}

	expectedLast := 100.0
	if !almostEqual(replenishments[2], expectedLast) {
		t.Errorf("esperado ultima reposicao=%.2f, obtido=%.2f", expectedLast, replenishments[2])
	}
}
