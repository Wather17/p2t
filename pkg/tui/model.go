package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type activeTab int

const (
	tabTelemetry activeTab = 0
	tabBuffer    activeTab = 1
	tabHistory   activeTab = 2
)

// MainModel e o modelo principal da TUI do p2t.
type MainModel struct {
	activeTab activeTab
	db        *sql.DB
	repo      *storage.Repository

	// Campos de entrada Telemetria
	inputSalary   textinput.Model
	inputContract textinput.Model
	inputCommute  textinput.Model
	inputErrors   textinput.Model
	inputCosts    textinput.Model
	activeField   int

	// Campos de entrada Buffer
	inputCap       textinput.Model
	inputRemaining textinput.Model
	activeBufferField int

	// Resultados
	telemetryResult *p2t.TelemetryResult
	bufferMetrics   *p2t.BufferMetrics
	historyTable    table.Model
	statusMsg       string
	width           int
	height          int
}

// NewMainModel inicializa o modelo TUI.
func NewMainModel(db *sql.DB) MainModel {
	repo := storage.NewRepository(db)

	m := MainModel{
		activeTab: tabTelemetry,
		db:        db,
		repo:      repo,
	}

	// Inicializa inputs da Telemetria
	m.inputSalary = textinput.New()
	m.inputSalary.Placeholder = "ex: 5000,00"
	m.inputSalary.Focus()

	m.inputContract = textinput.New()
	m.inputContract.Placeholder = "ex: 160"

	m.inputCommute = textinput.New()
	m.inputCommute.Placeholder = "ex: 40"

	m.inputErrors = textinput.New()
	m.inputErrors.Placeholder = "ex: 0,00"

	m.inputCosts = textinput.New()
	m.inputCosts.Placeholder = "ex: 200,00"

	// Inicializa inputs do Buffer
	m.inputCap = textinput.New()
	m.inputCap.Placeholder = "ex: 300,00"

	m.inputRemaining = textinput.New()
	m.inputRemaining.Placeholder = "ex: 120,00"

	// Inicializa tabela de historico
	m.initHistoryTable()

	return m
}

func (m *MainModel) initHistoryTable() {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Data", Width: 12},
		{Title: "Bruto (SB)", Width: 12},
		{Title: "VRH", Width: 12},
		{Title: "IDT", Width: 10},
	}

	var rows []table.Row
	if m.repo != nil {
		history, err := m.repo.GetRecentIDTHistory(10)
		if err == nil {
			for i, idt := range history {
				rows = append(rows, table.Row{
					fmt.Sprintf("%d", i+1),
					"Recente",
					"-",
					"-",
					fmt.Sprintf("%.2f%%", idt),
				})
			}
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	m.historyTable = t
}

func (m MainModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil
		case "shift+tab":
			if m.activeTab == 0 {
				m.activeTab = 2
			} else {
				m.activeTab--
			}
			return m, nil
		case "enter":
			m.calculateCurrentTab()
			return m, nil
		}
	}

	if m.activeTab == tabTelemetry {
		m.updateTelemetryInputs(msg)
	} else if m.activeTab == tabBuffer {
		m.updateBufferInputs(msg)
	} else if m.activeTab == tabHistory {
		m.historyTable, cmd = m.historyTable.Update(msg)
	}

	return m, cmd
}

func (m *MainModel) updateTelemetryInputs(msg tea.Msg) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "down":
			m.activeField = (m.activeField + 1) % 5
			m.focusTelemetryField()
			return
		case "up":
			if m.activeField == 0 {
				m.activeField = 4
			} else {
				m.activeField--
			}
			m.focusTelemetryField()
			return
		}
	}

	switch m.activeField {
	case 0:
		m.inputSalary, _ = m.inputSalary.Update(msg)
	case 1:
		m.inputContract, _ = m.inputContract.Update(msg)
	case 2:
		m.inputCommute, _ = m.inputCommute.Update(msg)
	case 3:
		m.inputErrors, _ = m.inputErrors.Update(msg)
	case 4:
		m.inputCosts, _ = m.inputCosts.Update(msg)
	}
}

func (m *MainModel) focusTelemetryField() {
	m.inputSalary.Blur()
	m.inputContract.Blur()
	m.inputCommute.Blur()
	m.inputErrors.Blur()
	m.inputCosts.Blur()

	switch m.activeField {
	case 0:
		m.inputSalary.Focus()
	case 1:
		m.inputContract.Focus()
	case 2:
		m.inputCommute.Focus()
	case 3:
		m.inputErrors.Focus()
	case 4:
		m.inputCosts.Focus()
	}
}

func (m *MainModel) updateBufferInputs(msg tea.Msg) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "down", "up":
			m.activeBufferField = (m.activeBufferField + 1) % 2
			if m.activeBufferField == 0 {
				m.inputCap.Focus()
				m.inputRemaining.Blur()
			} else {
				m.inputCap.Blur()
				m.inputRemaining.Focus()
			}
			return
		}
	}

	if m.activeBufferField == 0 {
		m.inputCap, _ = m.inputCap.Update(msg)
	} else {
		m.inputRemaining, _ = m.inputRemaining.Update(msg)
	}
}

func (m *MainModel) calculateCurrentTab() {
	if m.activeTab == tabTelemetry {
		sb, _ := p2t.ParseBrazilianFloat(m.inputSalary.Value())
		hc, _ := p2t.ParseBrazilianFloat(m.inputContract.Value())
		hd, _ := p2t.ParseBrazilianFloat(m.inputCommute.Value())
		de, _ := p2t.ParseBrazilianFloat(m.inputErrors.Value())
		ci, _ := p2t.ParseBrazilianFloat(m.inputCosts.Value())

		if sb <= 0 || hc <= 0 {
			m.statusMsg = "Erro: Salário Bruto e Carga Horária são obrigatórios."
			return
		}

		taxes := p2t.EstimateLegalDeductions(sb)
		input := p2t.TelemetryInput{
			GrossSalary:     sb,
			FixedDeductions: taxes.TotalDeductions,
			ErrorDeductions: de,
			InvisibleCosts:  ci,
			ContractHours:   hc,
			CommuteHours:    hd,
		}

		res, err := p2t.CalculateTelemetry(input)
		if err != nil {
			m.statusMsg = fmt.Sprintf("Erro: %v", err)
			return
		}

		m.telemetryResult = &res
		m.statusMsg = "Telemetria calculada com sucesso!"

		if m.repo != nil {
			_, _ = m.repo.SaveTelemetry(input, res)
			m.initHistoryTable()
		}
	} else if m.activeTab == tabBuffer {
		cap, _ := p2t.ParseBrazilianFloat(m.inputCap.Value())
		rem, _ := p2t.ParseBrazilianFloat(m.inputRemaining.Value())

		if cap <= 0 {
			m.statusMsg = "Erro: Teto do Buffer (T) é obrigatório."
			return
		}

		cycle := p2t.BufferCycle{Cap: cap, RemainingBalance: rem}
		ci, err := cycle.CalculateInvisibleCost()
		if err != nil {
			m.statusMsg = fmt.Sprintf("Erro: %v", err)
			return
		}

		metrics, err := p2t.CalculateBufferMetrics(cap, []float64{ci})
		if err != nil {
			m.statusMsg = fmt.Sprintf("Erro: %v", err)
			return
		}

		m.bufferMetrics = &metrics
		m.statusMsg = "Buffer Operacional calculado com sucesso!"

		if m.repo != nil {
			_, _ = m.repo.SaveBufferCycle(cap, rem, ci)
		}
	}
}

func (m MainModel) View() string {
	var b strings.Builder

	// Titulo do App
	b.WriteString(AppTitleStyle.Render("p2t - Pay to Work Framework (TUI)"))
	b.WriteString("\n\n")

	// Barra de Abas
	tabNames := []string{"1. Telemetria de Retorno", "2. Buffer Operacional", "3. Histórico SQLite"}
	for i, name := range tabNames {
		if activeTab(i) == m.activeTab {
			b.WriteString(ActiveTabStyle.Render(name))
		} else {
			b.WriteString(TabStyle.Render(name))
		}
	}
	b.WriteString("\n\n")

	// Conteudo da Aba Ativa
	switch m.activeTab {
	case tabTelemetry:
		b.WriteString(m.viewTelemetryTab())
	case tabBuffer:
		b.WriteString(m.viewBufferTab())
	case tabHistory:
		b.WriteString(m.viewHistoryTab())
	}

	// Status e Ajuda
	if m.statusMsg != "" {
		b.WriteString("\n" + CardStyle.Render(m.statusMsg))
	}

	b.WriteString(HelpStyle.Render("[Tab/Shift+Tab] Navegar Abas  |  [Seta Cima/Baixo] Mudar Campo  |  [Enter] Calcular  |  [q] Sair"))

	return b.String()
}

func (m MainModel) viewTelemetryTab() string {
	var b strings.Builder

	b.WriteString("--- Formulário de Entrada ---\n")
	b.WriteString(fmt.Sprintf("Salário Bruto (SB):     %s\n", m.inputSalary.View()))
	b.WriteString(fmt.Sprintf("Horas Contratuais (HC): %s\n", m.inputContract.View()))
	b.WriteString(fmt.Sprintf("Horas Deslocamento (HD):%s\n", m.inputCommute.View()))
	b.WriteString(fmt.Sprintf("Descontos por Erros (DE):%s\n", m.inputErrors.View()))
	b.WriteString(fmt.Sprintf("Custos Invisíveis (CI): %s\n", m.inputCosts.View()))

	if m.telemetryResult != nil {
		res := m.telemetryResult
		zone, desc := p2t.EvaluateIDTZone(res.IDT)

		var badge string
		switch zone {
		case p2t.ZoneGreen:
			badge = BadgeGreenStyle.Render(string(zone))
		case p2t.ZoneYellow:
			badge = BadgeYellowStyle.Render(string(zone))
		default:
			badge = BadgeRedStyle.Render(string(zone))
		}

		card := fmt.Sprintf(
			"Horas Totais (HT): %.2f h\nLiquidez Real (SL): R$ %.2f\nValor Real Hora (VRH): R$ %.2f/h\nDesperdício (IDT): %.2f%%\nDecisão: %s - %s",
			res.TotalHours, res.RealLiquidity, res.VRH, res.IDT, badge, desc,
		)
		b.WriteString("\n" + CardStyle.Render(card))
	}

	return b.String()
}

func (m MainModel) viewBufferTab() string {
	var b strings.Builder

	b.WriteString("--- Gestão do Buffer (A Caixinha) ---\n")
	b.WriteString(fmt.Sprintf("Teto Alocado (T):      %s\n", m.inputCap.View()))
	b.WriteString(fmt.Sprintf("Saldo Remanescente (S_rem): %s\n", m.inputRemaining.View()))

	if m.bufferMetrics != nil {
		diag, desc := p2t.DiagnoseEfficiency(m.bufferMetrics.TCM)
		card := fmt.Sprintf(
			"Média Reposição (R_bar): R$ %.2f\nTaxa Consumo (TCM): %.2f%%\nDiagnóstico: [%s] %s",
			m.bufferMetrics.AverageReplenishment, m.bufferMetrics.TCM, diag, desc,
		)
		b.WriteString("\n" + CardStyle.Render(card))
	}

	return b.String()
}

func (m MainModel) viewHistoryTab() string {
	var b strings.Builder
	b.WriteString("--- Registros Recentes (SQLite) ---\n")
	b.WriteString(m.historyTable.View())
	return b.String()
}
