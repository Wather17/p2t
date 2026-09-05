package cmd_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wather17/p2t/cmd"
)

func executeCommand(t *testing.T, args ...string) string {
	t.Helper()
	root := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("comando %v falhou: %v", args, err)
	}
	return buf.String()
}

func TestFinanceCommandsAndClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "finance.db")

	out := executeCommand(t, "goal", "add", "--db", dbPath,
		"--name", "Reserva", "--target", "10000", "--deadline", "2027-12",
		"--balance", "2500", "--monthly-contribution", "500")
	if !strings.Contains(out, "Meta #1 criada") {
		t.Fatalf("saida inesperada ao criar meta: %s", out)
	}
	out = executeCommand(t, "goal", "update", "--db", dbPath, "--id", "1", "--balance", "3000")
	if !strings.Contains(out, "Meta #1 atualizada") {
		t.Fatalf("saida inesperada ao atualizar meta: %s", out)
	}

	out = executeCommand(t, "recurring", "add", "--db", dbPath,
		"--name", "Streaming", "--amount", "50", "--period", "monthly",
		"--purpose", "life", "--essentiality", "discretionary", "--billing-day", "10")
	if !strings.Contains(out, "Compromisso #1 criado") {
		t.Fatalf("saida inesperada ao criar compromisso: %s", out)
	}
	out = executeCommand(t, "recurring", "update", "--db", dbPath, "--id", "1", "--amount", "60")
	if !strings.Contains(out, "Compromisso #1 atualizado") {
		t.Fatalf("saida inesperada ao atualizar compromisso: %s", out)
	}

	out = executeCommand(t, "close", "--db", dbPath, "--reference-month", "2026-08",
		"--received-income", "5000", "--reserve-balance", "2500",
		"--goal-contributions", "500", "--exceptional-costs", "100")
	if !strings.Contains(out, "Compromissos Recorrentes: R$ 60.00 (1 itens)") {
		t.Fatalf("compromisso não apareceu no fechamento: %s", out)
	}
	if !strings.Contains(out, "Capacidade de Metas: R$ 4940.00") {
		t.Fatalf("capacidade de metas inesperada: %s", out)
	}
	if !strings.Contains(out, "Margem Livre: R$ 4340.00") {
		t.Fatalf("margem livre inesperada: %s", out)
	}
	if !strings.Contains(out, "nenhum registro encontrado") {
		t.Fatalf("fechamento sem telemetria deveria ser explícito: %s", out)
	}

	out = executeCommand(t, "close", "--db", dbPath, "--reference-month", "2026-08",
		"--received-income", "5100", "--reserve-balance", "2600",
		"--goal-contributions", "600", "--exceptional-costs", "0")
	if !strings.Contains(out, "Renda Recebida: R$ 5100.00") {
		t.Fatalf("segundo fechamento não atualizou a competência: %s", out)
	}

	out = executeCommand(t, "goal", "archive", "--db", dbPath, "--id", "1")
	if !strings.Contains(out, "Meta #1 arquivada") {
		t.Fatalf("saida inesperada ao arquivar meta: %s", out)
	}
	out = executeCommand(t, "recurring", "archive", "--db", dbPath, "--id", "1")
	if !strings.Contains(out, "Compromisso #1 arquivado") {
		t.Fatalf("saida inesperada ao arquivar compromisso: %s", out)
	}
}

func TestGoalUpdateRequiresAField(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "finance.db")
	executeCommand(t, "goal", "add", "--db", dbPath,
		"--name", "Reserva", "--target", "1000", "--deadline", "2027-12")

	root := cmd.NewRootCmd()
	root.SetArgs([]string{"goal", "update", "--db", dbPath, "--id", "1"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "ao menos um campo") {
		t.Fatalf("esperado erro de atualização sem campos, obtido: %v", err)
	}
}

func TestCloseInteractive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "finance.db")
	root := cmd.NewRootCmd()
	out := new(bytes.Buffer)
	root.SetOut(out)
	root.SetIn(strings.NewReader("5000\n2500\n500\n100\n"))
	root.SetArgs([]string{
		"close", "--interactive", "--db", dbPath, "--reference-month", "2026-09",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("fechamento interativo falhou: %v", err)
	}
	if !strings.Contains(out.String(), "Renda Recebida: R$ 5000.00") {
		t.Fatalf("renda do fluxo interativo não apareceu: %s", out.String())
	}
}
