package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wather17/p2t/cmd"
	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
)

func TestExportAndImportCLI_JSON(t *testing.T) {
	tempDir := t.TempDir()
	dbFile1 := filepath.Join(tempDir, "db1.db")
	exportFile := filepath.Join(tempDir, "export.json")
	dbFile2 := filepath.Join(tempDir, "db2.db")

	// 1. Criar dados no db1
	db1, err := storage.OpenDB(dbFile1)
	if err != nil {
		t.Fatalf("falha ao abrir db1: %v", err)
	}
	repo1 := storage.NewRepository(db1)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ContractHours:   160.0,
	}
	res, _ := p2t.CalculateTelemetry(input)
	_, _ = repo1.SaveTelemetry(input, res, "2026-01")
	db1.Close()

	// 2. Exportar via CLI
	rootCmdExport := cmd.NewRootCmd()
	bufExport := new(bytes.Buffer)
	rootCmdExport.SetOut(bufExport)
	rootCmdExport.SetArgs([]string{
		"export",
		"--format", "json",
		"--output", exportFile,
		"--db", dbFile1,
	})
	if err := rootCmdExport.Execute(); err != nil {
		t.Fatalf("erro ao executar p2t export: %v", err)
	}

	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Fatalf("arquivo exportado %s nao existe", exportFile)
	}

	// 3. Importar via CLI no db2
	rootCmdImport := cmd.NewRootCmd()
	bufImport := new(bytes.Buffer)
	rootCmdImport.SetOut(bufImport)
	rootCmdImport.SetArgs([]string{
		"import",
		"--file", exportFile,
		"--db", dbFile2,
	})
	if err := rootCmdImport.Execute(); err != nil {
		t.Fatalf("erro ao executar p2t import: %v", err)
	}

	outImport := bufImport.String()
	if !strings.Contains(outImport, "1 registro(s) inserido(s)") {
		t.Errorf("saida inesperada no p2t import: %s", outImport)
	}
}

func TestExportCLI_Stdout(t *testing.T) {
	tempDir := t.TempDir()
	dbFile := filepath.Join(tempDir, "db.db")

	db, err := storage.OpenDB(dbFile)
	if err != nil {
		t.Fatalf("falha ao abrir db: %v", err)
	}
	repo := storage.NewRepository(db)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ContractHours:   160.0,
	}
	res, _ := p2t.CalculateTelemetry(input)
	_, _ = repo.SaveTelemetry(input, res, "2026-02")
	db.Close()

	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"export",
		"--format", "csv",
		"--db", dbFile,
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("erro ao executar p2t export para stdout: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "reference_month") || !strings.Contains(out, "2026-02") {
		t.Errorf("saida CSV esperada em stdout nao encontrada: %s", out)
	}
}

func TestImportCLI_WithFailures(t *testing.T) {
	tempDir := t.TempDir()
	dbFile := filepath.Join(tempDir, "db.db")
	importFile := filepath.Join(tempDir, "import.json")

	data := []byte(`[
		{"ID":1,"ReferenceMonth":"2026-05","GrossSalary":5000,"FixedDeductions":800,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-05-01T10:00:00Z"},
		{"ID":2,"ReferenceMonth":"2026-05","GrossSalary":5000,"FixedDeductions":800,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-05-01T10:00:00Z"},
		{"ID":3,"ReferenceMonth":"2026-06","GrossSalary":0,"FixedDeductions":0,"ErrorDeductions":0,"InvisibleCosts":0,"ContractHours":160,"CommuteHours":40,"CreatedAt":"2026-06-01T10:00:00Z"}
	]`)
	if err := os.WriteFile(importFile, data, 0o600); err != nil {
		t.Fatalf("falha ao gravar arquivo de importacao: %v", err)
	}

	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"import", "-F", importFile, "--db", dbFile})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("erro ao executar import: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1 registro(s) inserido(s)") {
		t.Errorf("esperado resumo de inseridos: %s", out)
	}
	if !strings.Contains(out, "1 duplicado(s) ignorado(s)") {
		t.Errorf("esperado resumo de duplicados: %s", out)
	}
	if !strings.Contains(out, "1 falha(s)") {
		t.Errorf("esperado resumo de falhas: %s", out)
	}
}
