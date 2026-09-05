package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Wather17/p2t/cmd"
)

func TestMathCmd_Success(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"math"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando math: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "p2t - Organização Financeira & Método Dedutivo") {
		t.Errorf("titulo do comando math nao encontrado: %s", out)
	}
	if !strings.Contains(out, "GLOSSÁRIO DE INCÓGNITAS") {
		t.Errorf("glossario nao encontrado na saida do comando math: %s", out)
	}
	if !strings.Contains(out, "EQUAÇÕES FUNDAMENTAIS") {
		t.Errorf("equacoes nao encontradas na saida do comando math: %s", out)
	}
	if !strings.Contains(out, "ORGANIZAÇÃO FINANCEIRA POR EXCEÇÃO") {
		t.Errorf("modelo financeiro nao encontrado na saida do comando math: %s", out)
	}
}

func TestMathCmd_AliasDocs(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"docs"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar alias docs: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "p2t - Organização Financeira & Método Dedutivo") {
		t.Errorf("titulo do alias docs nao encontrado: %s", out)
	}
}
