package storage_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
)

func TestExportAndImportJSON(t *testing.T) {
	db1, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir db1: %v", err)
	}
	defer db1.Close()

	repo1 := storage.NewRepository(db1)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ContractHours:   160.0,
	}
	res, _ := p2t.CalculateTelemetry(input)
	_, err = repo1.SaveTelemetry(input, res, "2026-05")
	if err != nil {
		t.Fatalf("falha ao salvar registro no db1: %v", err)
	}

	// Exportar para JSON
	jsonData, err := storage.ExportTelemetryJSON(repo1)
	if err != nil {
		t.Fatalf("falha ao exportar JSON: %v", err)
	}

	// Importar em novo banco
	db2, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir db2: %v", err)
	}
	defer db2.Close()

	repo2 := storage.NewRepository(db2)
	importedCount, err := storage.ImportTelemetryJSON(repo2, jsonData)
	if err != nil {
		t.Fatalf("falha ao importar JSON: %v", err)
	}

	if importedCount != 1 {
		t.Errorf("esperado 1 registro importado, obtido: %d", importedCount)
	}

	rec, err := repo2.GetTelemetryByReferenceMonth("2026-05")
	if err != nil || rec == nil {
		t.Fatalf("registro nao encontrado apos importacao JSON: %v", err)
	}
}

func TestExportAndImportCSV(t *testing.T) {
	db1, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir db1: %v", err)
	}
	defer db1.Close()

	repo1 := storage.NewRepository(db1)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ContractHours:   160.0,
	}
	res, _ := p2t.CalculateTelemetry(input)
	_, err = repo1.SaveTelemetry(input, res, "2026-06")
	if err != nil {
		t.Fatalf("falha ao salvar registro no db1: %v", err)
	}

	// Exportar para CSV
	csvData, err := storage.ExportTelemetryCSV(repo1)
	if err != nil {
		t.Fatalf("falha ao exportar CSV: %v", err)
	}

	// Importar em novo banco
	db2, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir db2: %v", err)
	}
	defer db2.Close()

	repo2 := storage.NewRepository(db2)
	importedCount, err := storage.ImportTelemetryCSV(repo2, csvData)
	if err != nil {
		t.Fatalf("falha ao importar CSV: %v", err)
	}

	if importedCount != 1 {
		t.Errorf("esperado 1 registro importado, obtido: %d", importedCount)
	}

	rec, err := repo2.GetTelemetryByReferenceMonth("2026-06")
	if err != nil || rec == nil {
		t.Fatalf("registro nao encontrado apos importacao CSV: %v", err)
	}
}

func TestExportPaginationNoTruncation(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ContractHours:   160.0,
	}
	res, _ := p2t.CalculateTelemetry(input)

	const total = 1200
	for i := 0; i < total; i++ {
		if _, err := repo.SaveTelemetry(input, res, ""); err != nil {
			t.Fatalf("falha ao salvar registro %d: %v", i, err)
		}
	}

	jsonData, err := storage.ExportTelemetryJSON(repo)
	if err != nil {
		t.Fatalf("falha ao exportar JSON: %v", err)
	}
	var records []storage.TelemetryRecord
	if err := json.Unmarshal(jsonData, &records); err != nil {
		t.Fatalf("falha ao parsear JSON exportado: %v", err)
	}
	if len(records) != total {
		t.Errorf("JSON exportado esperado com %d registros, obtido %d", total, len(records))
	}

	csvData, err := storage.ExportTelemetryCSV(repo)
	if err != nil {
		t.Fatalf("falha ao exportar CSV: %v", err)
	}
	reader := csv.NewReader(bytes.NewReader(csvData))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("falha ao parsear CSV exportado: %v", err)
	}
	if len(rows) != total+1 {
		t.Errorf("CSV esperado com %d linhas (cabecalho + registros), obtido %d", total+1, len(rows))
	}
}
