package tui_test

import (
	"strings"
	"testing"

	"github.com/Wather17/p2t/pkg/storage"
	"github.com/Wather17/p2t/pkg/tui"
)

func TestMainModel_Render(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	model := tui.NewMainModel(db)
	view := model.View()

	if !strings.Contains(view, "p2t - Pay to Work Framework (TUI)") {
		t.Errorf("titulo da TUI nao encontrado na renderizacao: %s", view)
	}
	if !strings.Contains(view, "1. Telemetria de Retorno") {
		t.Errorf("aba de telemetria nao encontrada na renderizacao: %s", view)
	}
}
