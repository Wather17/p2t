package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Wather17/p2t/cmd"
)

func TestVersionCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando version: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "p2t CLI v") {
		t.Errorf("saida inesperada do comando version: %s", out)
	}
}

func TestTelemetryCmd_Success(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-s", "5000",
		"-f", "800",
		"-e", "200",
		"-c", "300",
		"-H", "160",
		"-d", "40",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando telemetry: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Carga Horaria Total (HT): 200.00 h") {
		t.Errorf("saida esperada HT nao encontrada: %s", out)
	}
	if !strings.Contains(out, "Valor Real da Hora (VRH): R$ 18.50 / h") {
		t.Errorf("saida esperada VRH nao encontrada: %s", out)
	}
	if !strings.Contains(out, "Indice de Desperdicio (IDT): 10.00%") {
		t.Errorf("saida esperada IDT nao encontrada: %s", out)
	}
}

func TestTelemetryCmd_WithSchedule(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-s", "5000",
		"-f", "800",
		"-H", "160",
		"-W", "5x2",
		"-D", "1.5",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando telemetry com schedule: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Escala de Trabalho: 5x2 (HD calculado: 33.00 h/mês)") {
		t.Errorf("saida esperada de escala 5x2 nao encontrada: %s", out)
	}
}

func TestTelemetryCmd_WithExactShiftRefDate(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-s", "5000",
		"-f", "800",
		"-H", "180",
		"-W", "12x36",
		"-D", "2.0",
		"-R", "2026-07-01",
		"-m", "2026-07",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando telemetry com refDate: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Escala de Trabalho: 12x36 (Calendário Exato: 16 plantões, HD: 32.00 h/mês)") {
		t.Errorf("saida esperada de escala 12x36 exata nao encontrada: %s", out)
	}
}

func TestTelemetryCmd_RefDateRejectedForNon12x36(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-s", "5000",
		"-f", "800",
		"-H", "160",
		"-W", "5x2",
		"-D", "1.5",
		"-R", "2026-08-05",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("esperado erro ao usar -R com escala diferente de 12x36")
	}
	if !strings.Contains(err.Error(), "--shift-ref-date é exclusivo da escala 12x36") {
		t.Errorf("mensagem de erro inesperada: %v", err)
	}
	if strings.Contains(buf.String(), "Carga Horaria Total") {
		t.Errorf("nenhuma linha de resultado deveria ser impressa: %s", buf.String())
	}
}

func TestTelemetryCmd_RefDateRequiresSchedule(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-s", "5000",
		"-f", "800",
		"-H", "160",
		"-R", "2026-08-05",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("esperado erro ao usar -R sem -W 12x36")
	}
	if !strings.Contains(err.Error(), "--shift-ref-date é exclusivo da escala 12x36") {
		t.Errorf("mensagem de erro inesperada: %v", err)
	}
}

func TestTelemetryCmd_RefDateValidFor12x36(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-s", "5000",
		"-f", "800",
		"-H", "180",
		"-W", "12x36",
		"-D", "2.0",
		"-R", "2026-07-01",
		"-m", "2026-07",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando telemetry 12x36 com refDate valida: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Calendário Exato: 16 plantões") {
		t.Errorf("saida esperada de calendario exato nao encontrada: %s", out)
	}
}

func TestTelemetryCmd_InteractiveBrazilianFormat(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	outBuf := new(bytes.Buffer)
	inInput := "R$ 5.000,00\n160,0\n40\n800,00\n200,00\n300,00\n"
	inBuf := bytes.NewBufferString(inInput)

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetArgs([]string{
		"telemetry",
		"--no-save",
		"-I",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro no modo interativo do telemetry: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, "=== Modo Interativo: Telemetria p2t ===") {
		t.Errorf("cabecalho do modo interativo nao encontrado: %s", out)
	}
	if !strings.Contains(out, "Valor Real da Hora (VRH): R$ 18.50 / h") {
		t.Errorf("VRH invalido na saida interativa: %s", out)
	}
}

func TestBufferCmd_Success(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"buffer",
		"--no-save",
		"-t", "300",
		"-r", "180",
		"-p", "100,120,140",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando buffer: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Teto Alocado (T): R$ 300.00") {
		t.Errorf("saida esperada Teto nao encontrada: %s", out)
	}
	if !strings.Contains(out, "Custo Invisivel do Ciclo (CI / Rt): R$ 120.00") {
		t.Errorf("saida esperada CI nao encontrada: %s", out)
	}
	if !strings.Contains(out, "Taxa de Consumo Media (TCM): 40.00%") {
		t.Errorf("saida esperada TCM nao encontrada: %s", out)
	}
}

func TestHistoryCmd_Success(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{
		"history",
		"--db", ":memory:",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando history: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "=== Histórico de Telemetria") {
		t.Errorf("saida esperada do history nao encontrada: %s", out)
	}
}
