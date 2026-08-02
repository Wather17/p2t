package tui_test

import (
	"strings"
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/Wather17/p2t/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
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

func TestMainModel_PrefillAndTabSwitch(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	input := p2t.TelemetryInput{
		GrossSalary:     6000.0,
		FixedDeductions: 1000.0,
		ErrorDeductions: 100.0,
		InvisibleCosts:  200.0,
		ContractHours:   160.0,
		CommuteHours:    30.0,
	}
	res, _ := p2t.CalculateTelemetry(input)
	_, err = repo.SaveTelemetry(input, res, "2026-06")
	if err != nil {
		t.Fatalf("falha ao salvar registro no banco: %v", err)
	}

	model := tui.NewMainModel(db)

	// Testar troca de abas via KeyMsg '2' (Buffer Operacional)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	viewBuffer := updatedModel.View()
	if !strings.Contains(viewBuffer, "Gestão do Buffer Operacional") {
		t.Errorf("aba buffer nao foi ativada com atalho '2': %s", viewBuffer)
	}

	// Testar troca de abas via KeyMsg '3' (Historico SQLite)
	updatedModel2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	viewHistory := updatedModel2.View()
	if !strings.Contains(viewHistory, "Histórico de Telemetria (SQLite)") {
		t.Errorf("aba historico nao foi ativada com atalho '3': %s", viewHistory)
	}
}

func TestRenderProgressBar(t *testing.T) {
	bar := tui.RenderProgressBar(50.0, 20, tui.GreenZoneColor)
	if len(bar) == 0 {
		t.Errorf("esperado string nao vazia para progress bar")
	}
}

