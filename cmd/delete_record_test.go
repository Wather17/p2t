package cmd_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Wather17/p2t/cmd"
	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
)

func TestDeleteRecordCmd_Success(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test.db")
	db, err := storage.OpenDB(dbFile)
	if err != nil {
		t.Fatalf("falha ao abrir banco em arquivo temporario: %v", err)
	}

	repo := storage.NewRepository(db)
	input := p2t.TelemetryInput{
		GrossSalary:     5000,
		FixedDeductions: 800,
		ContractHours:   160,
	}
	res, _ := p2t.CalculateTelemetry(input)
	id, err := repo.SaveTelemetry(input, res, "2026-05")
	if err != nil || id <= 0 {
		t.Fatalf("falha ao salvar registro: %v", err)
	}
	db.Close()

	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"delete-record",
		"--id", fmt.Sprintf("%d", id),
		"--db", dbFile,
	})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar delete-record: %v", err)
	}
}

func TestDeleteRecordCmd_InvalidID(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"delete-record",
		"--id", "0",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("esperado erro ao passar ID zero")
	}
}
