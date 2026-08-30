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

func typeValue(t *testing.T, m tui.MainModel, value string) tui.MainModel {
	t.Helper()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)})
	model, ok := updated.(tui.MainModel)
	if !ok {
		t.Fatalf("Update retornou tipo inesperado: %T", updated)
	}
	return model
}

func pressKey(t *testing.T, m tui.MainModel, key tea.KeyMsg) tui.MainModel {
	t.Helper()
	updated, _ := m.Update(key)
	model, ok := updated.(tui.MainModel)
	if !ok {
		t.Fatalf("Update retornou tipo inesperado: %T", updated)
	}
	return model
}

func fillTelemetryInputs(t *testing.T, m tui.MainModel) tui.MainModel {
	t.Helper()
	m = typeValue(t, m, "5000")
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	m = typeValue(t, m, "460")
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	m = typeValue(t, m, "40")
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	return m
}

func TestMainModel_TelemetrySaveStatus(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	model := tui.NewMainModel(db)
	model = fillTelemetryInputs(t, model)

	model = pressKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(model.View(), "salva com sucesso") {
		t.Errorf("esperado status de sucesso apos primeiro save: %s", model.View())
	}

	records, err := repo.GetRecentTelemetryRecords(15)
	if err != nil {
		t.Fatalf("falha ao consultar registros: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("esperado 1 registro salvo, obtido %d", len(records))
	}

	// Segunda execucao com o mesmo mes de competencia -> duplicado
	model = pressKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(model.View(), "já existia") {
		t.Errorf("esperado status de duplicidade apos segundo save: %s", model.View())
	}

	records, err = repo.GetRecentTelemetryRecords(15)
	if err != nil {
		t.Fatalf("falha ao consultar registros: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("esperado 1 registro apos duplicado (nao sobrescrever), obtido %d", len(records))
	}
}

func TestMainModel_TelemetrySaveFailsWithoutDB(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}

	model := tui.NewMainModel(db)
	model = fillTelemetryInputs(t, model)

	if err := db.Close(); err != nil {
		t.Fatalf("falha ao fechar banco: %v", err)
	}

	model = pressKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	view := model.View()
	if strings.Contains(view, "salva com sucesso") {
		t.Errorf("status nao pode afirmar sucesso com banco indisponivel: %s", view)
	}
	if !strings.Contains(view, "não foi possível salvar") {
		t.Errorf("esperado status explicando falha de persistencia: %s", view)
	}
}

func TestMainModel_CommutePrefillFractional(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 1000.0,
		ErrorDeductions: 100.0,
		InvisibleCosts:  200.0,
		ContractHours:   160.0,
		CommuteHours:    31.5,
	}
	res, err := p2t.CalculateTelemetry(input)
	if err != nil {
		t.Fatalf("falha ao calcular telemetria: %v", err)
	}
	if _, err := repo.SaveTelemetry(input, res, "2026-06"); err != nil {
		t.Fatalf("falha ao salvar registro: %v", err)
	}

	model := tui.NewMainModel(db)
	if !strings.Contains(model.View(), "31.50") {
		t.Errorf("esperado prefill de deslocamento com 2 casas decimais (%s) na View(): %s", "31.50", model.View())
	}

	// Enter sem edicao re-salva com o mesmo valor exato
	model = pressKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	latest, err := repo.GetLatestTelemetryRecord()
	if err != nil {
		t.Fatalf("falha ao obter registro mais recente: %v", err)
	}
	if latest == nil {
		t.Fatalf("esperado registro mais recente, obtido nil")
	}
	if latest.CommuteHours != 31.5 {
		t.Errorf("esperado commute_hours = 31.5 apos re-save, obtido %.2f", latest.CommuteHours)
	}
}

