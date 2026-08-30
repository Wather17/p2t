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

	if importedCount.Inserted != 1 {
		t.Errorf("esperado 1 registro importado, obtido: %d", importedCount.Inserted)
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

	if importedCount.Inserted != 1 {
		t.Errorf("esperado 1 registro importado, obtido: %d", importedCount.Inserted)
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

func TestImportJSON_TripartiteResult(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ContractHours:   160.0,
	}
	res, _ := p2t.CalculateTelemetry(input)
	if _, err := repo.SaveTelemetry(input, res, "2026-01"); err != nil {
		t.Fatalf("falha ao salvar registro pre-existente: %v", err)
	}

	data := []byte(`[
		{"ID":10,"ReferenceMonth":"2026-02","GrossSalary":5000,"FixedDeductions":800,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-02-01T10:00:00Z"},
		{"ID":11,"ReferenceMonth":"2026-01","GrossSalary":5000,"FixedDeductions":800,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-01-01T10:00:00Z"},
		{"ID":12,"ReferenceMonth":"2026-03","GrossSalary":0,"FixedDeductions":0,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-03-01T10:00:00Z"}
	]`)

	result, err := storage.ImportTelemetryJSON(repo, data)
	if err != nil {
		t.Fatalf("falha ao importar JSON: %v", err)
	}
	if result.Inserted != 1 || result.Duplicated != 1 || result.Failed != 1 {
		t.Errorf("esperado 1/1/1, obtido inseridos=%d duplicados=%d falhas=%d", result.Inserted, result.Duplicated, result.Failed)
	}
	if len(result.Why) != 2 {
		t.Errorf("esperado 2 motivos, obtido %d: %v", len(result.Why), result.Why)
	}
}

func TestImportJSON_ReimportAllDuplicated(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	data := []byte(`[
		{"ID":1,"ReferenceMonth":"2026-05","GrossSalary":5000,"FixedDeductions":800,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-05-01T10:00:00Z"}
	]`)

	first, err := storage.ImportTelemetryJSON(repo, data)
	if err != nil {
		t.Fatalf("falha na primeira importacao: %v", err)
	}
	if first.Inserted != 1 {
		t.Fatalf("esperado 1 inserido na primeira passada, obtido %d", first.Inserted)
	}

	second, err := storage.ImportTelemetryJSON(repo, data)
	if err != nil {
		t.Fatalf("falha na segunda importacao: %v", err)
	}
	if second.Inserted != 0 || second.Duplicated != 1 || second.Failed != 0 {
		t.Errorf("reimport esperado 0 inseridos/1 duplicado, obtido %+v", second)
	}
	if len(second.Why) == 0 {
		t.Errorf("esperado motivo explicando duplicidade no reimport")
	}
}

func TestImportCSV_BadRowsAccounting(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	header := "id,reference_month,gross_salary,fixed_deductions,error_deductions,invisible_costs,contract_hours,commute_hours,total_hours,real_liquidity,vrh,idt,created_at\n"
	shortRow := "1,,5000\n"
	badNumRow := "2,2026-02,abc,800,0,0,160,40,200,4200,26.25,0,2026-02-01 10:00:00\n"
	validRow := "3,2026-03,5000,800,0,0,160,40,200,4200,26.25,0,2026-03-01 10:00:00\n"
	data := []byte(header + shortRow + badNumRow + validRow)

	result, err := storage.ImportTelemetryCSV(repo, data)
	if err != nil {
		t.Fatalf("falha ao importar CSV: %v", err)
	}
	if result.Inserted != 1 || result.Failed != 2 {
		t.Errorf("esperado 1 inserido/2 falhas, obtido %+v", result)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM telemetry_cycles").Scan(&count); err != nil {
		t.Fatalf("falha ao contar registros: %v", err)
	}
	if count != 1 {
		t.Errorf("banco deveria conter apenas o registro valido; obtido %d", count)
	}
}
