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

func TestRepository_TelemetryChunks(t *testing.T) {
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

	const total = 1200
	for i := 0; i < total; i++ {
		if _, err := repo.SaveTelemetry(input, res, ""); err != nil {
			t.Fatalf("falha ao salvar registro %d: %v", i, err)
		}
	}

	firstChunk, err := repo.GetTelemetryRecordsChunk(0, 1000)
	if err != nil {
		t.Fatalf("falha ao buscar primeiro chunk: %v", err)
	}
	if len(firstChunk) != 1000 {
		t.Fatalf("primeiro chunk esperado com 1000 registros, obtido %d", len(firstChunk))
	}

	lastID := firstChunk[len(firstChunk)-1].ID
	secondChunk, err := repo.GetTelemetryRecordsChunk(lastID, 1000)
	if err != nil {
		t.Fatalf("falha ao buscar segundo chunk: %v", err)
	}
	if len(secondChunk) != total-1000 {
		t.Fatalf("segundo chunk esperado com %d registros, obtido %d", total-1000, len(secondChunk))
	}

	all := append(firstChunk, secondChunk...)
	for i := 1; i < len(all); i++ {
		if all[i].ID <= all[i-1].ID {
			t.Errorf("ordem ascendente por id violada no indice %d (%d > %d)", i, all[i-1].ID, all[i].ID)
		}
	}
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

	id, err := repo.SaveTelemetry(input, res, "2026-05")
	if err != nil {
		t.Fatalf("falha ao salvar telemetria no banco: %v", err)
	}
	if id <= 0 {
		t.Errorf("esperado ID > 0, obtido: %d", id)
	}

	// Adiciona mais dois ciclos para testar historico de IDT
	_, _ = repo.SaveTelemetry(input, res, "2026-06")
	input2 := input
	input2.ErrorDeductions = 500.0 // altera IDT
	res2, _ := p2t.CalculateTelemetry(input2)
	_, _ = repo.SaveTelemetry(input2, res2, "2026-07")

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

	records, err := repo.GetRecentTelemetryRecords(3)
	if err != nil {
		t.Fatalf("falha ao consultar registros de telemetria: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("esperado 3 registros completos, obtido: %d", len(records))
	}

	latest, err := repo.GetLatestTelemetryRecord()
	if err != nil {
		t.Fatalf("falha ao consultar ultimo registro de telemetria: %v", err)
	}
	if latest == nil || !almostEqual(latest.IDT, res2.IDT) {
		t.Errorf("esperado ultimo registro com IDT=%.2f, obtido: %v", res2.IDT, latest)
	}

	recByMonth, err := repo.GetTelemetryByReferenceMonth("2026-07")
	if err != nil || recByMonth == nil {
		t.Fatalf("falha ao consultar telemetria por competencia 2026-07: %v", err)
	}
	if recByMonth.ReferenceMonth != "2026-07" {
		t.Errorf("esperado competencia '2026-07', obtido: '%s'", recByMonth.ReferenceMonth)
	}

	// Testar trava de duplicidade para o mesmo mes de competencia
	_, errDup := repo.SaveTelemetry(input, res, "2026-07")
	if errDup == nil {
		t.Errorf("esperado erro de duplicidade ao salvar segundo registro com a mesma competencia '2026-07'")
	}
}

func TestRepository_Buffer(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	id, err := repo.SaveBufferCycle(300.0, 180.0, 120.0, "2026-06")
	if err != nil {
		t.Fatalf("falha ao salvar ciclo de buffer: %v", err)
	}
	if id <= 0 {
		t.Errorf("esperado ID > 0, obtido: %d", id)
	}

	_, _ = repo.SaveBufferCycle(300.0, 180.0, 120.0, "2026-05")
	_, _ = repo.SaveBufferCycle(300.0, 150.0, 150.0, "2026-06")
	_, _ = repo.SaveBufferCycle(300.0, 200.0, 100.0, "2026-07")

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

func TestRepository_GetTelemetryAnalytics(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	// Inserir 4 ciclos de teste
	months := []string{"2026-01", "2026-02", "2026-03", "2026-04"}
	idts := []float64{8.0, 10.0, 12.0, 14.0}

	for i, m := range months {
		input := p2t.TelemetryInput{
			GrossSalary:     5000.0,
			FixedDeductions: 800.0,
			ErrorDeductions: idts[i] * 50.0,
			InvisibleCosts:  100.0,
			ContractHours:   160.0,
			CommuteHours:    30.0,
		}
		res, _ := p2t.CalculateTelemetry(input)
		_, _ = repo.SaveTelemetry(input, res, m)
	}

	analytics, err := repo.GetTelemetryAnalytics()
	if err != nil {
		t.Fatalf("falha ao calcular analytics: %v", err)
	}

	if analytics.TotalRecords != 4 {
		t.Errorf("esperado 4 registros totais, obtido: %d", analytics.TotalRecords)
	}

	if analytics.IDT3 <= 0 {
		t.Errorf("esperado IDT3 > 0, obtido: %.2f", analytics.IDT3)
	}

	if analytics.TotalAnnualInvisibleCosts <= 0 {
		t.Errorf("esperado TotalAnnualInvisibleCosts > 0, obtido: %.2f", analytics.TotalAnnualInvisibleCosts)
	}
}

func TestRepository_DeleteAndUpdate(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ErrorDeductions: 0.0,
		InvisibleCosts:  100.0,
		ContractHours:   160.0,
		CommuteHours:    30.0,
	}
	res, _ := p2t.CalculateTelemetry(input)

	id, err := repo.SaveTelemetry(input, res, "2026-05")
	if err != nil {
		t.Fatalf("falha ao salvar registro: %v", err)
	}

	// Testar atualizacao
	input.GrossSalary = 6000.0
	resUpdated, _ := p2t.CalculateTelemetry(input)
	err = repo.UpdateTelemetryRecord(id, input, resUpdated, "2026-05")
	if err != nil {
		t.Fatalf("falha ao atualizar registro #%d: %v", id, err)
	}

	latest, _ := repo.GetLatestTelemetryRecord()
	if !almostEqual(latest.GrossSalary, 6000.0) {
		t.Errorf("esperado salario atualizado=6000.0, obtido=%.2f", latest.GrossSalary)
	}

	// Testar exclusao de telemetria
	err = repo.DeleteTelemetryRecord(id)
	if err != nil {
		t.Fatalf("falha ao excluir registro #%d: %v", id, err)
	}

	// Testar exclusao de buffer
	bufID, _ := repo.SaveBufferCycle(300.0, 150.0, 150.0, "2026-06")
	err = repo.DeleteBufferRecord(bufID)
	if err != nil {
		t.Fatalf("falha ao excluir registro de buffer #%d: %v", bufID, err)
	}
}

func TestRepository_BufferUpsertByReferenceMonth(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	if _, err := repo.SaveBufferCycle(300.0, 180.0, 120.0, "2026-06"); err != nil {
		t.Fatalf("falha ao salvar primeiro ciclo: %v", err)
	}
	if _, err := repo.SaveBufferCycle(300.0, 100.0, 200.0, "2026-06"); err != nil {
		t.Fatalf("falha ao salvar ciclo da mesma competencia: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM buffer_cycles").Scan(&count); err != nil {
		t.Fatalf("falha ao contar ciclos: %v", err)
	}
	if count != 1 {
		t.Errorf("esperado 1 ciclo apos dois saves da mesma competencia, obtido %d", count)
	}

	repls, err := repo.GetRecentReplenishments(10)
	if err != nil {
		t.Fatalf("falha ao consultar reposicoes: %v", err)
	}
	if len(repls) != 1 || !almostEqual(repls[0], 200.0) {
		t.Errorf("esperado reposicao 200.0 (valor substituido), obtido %v", repls)
	}

	if _, err := repo.SaveBufferCycle(300.0, 150.0, 150.0, "2026-07"); err != nil {
		t.Fatalf("falha ao salvar ciclo de outra competencia: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM buffer_cycles").Scan(&count); err != nil {
		t.Fatalf("falha ao contar ciclos: %v", err)
	}
	if count != 2 {
		t.Errorf("esperado 2 ciclos com competencias diferentes, obtido %d", count)
	}

	if _, err := repo.SaveBufferCycle(300.0, 180.0, 120.0, ""); err != nil {
		t.Fatalf("falha ao salvar ciclo legado vazio: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM buffer_cycles").Scan(&count); err != nil {
		t.Fatalf("falha ao contar ciclos: %v", err)
	}
	if count != 3 {
		t.Errorf("registros legados ('') nao conflitam; esperado 3 ciclos, obtido %d", count)
	}
}
