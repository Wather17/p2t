package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	inputCap          textinput.Model
	inputRemaining    textinput.Model
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

	// Tenta auto-preencher inputs com o registro mais recente do banco
	if repo != nil {
		if latest, err := repo.GetLatestTelemetryRecord(); err == nil && latest != nil {
			m.inputSalary.SetValue(fmt.Sprintf("%.2f", latest.GrossSalary))
			m.inputContract.SetValue(fmt.Sprintf("%.0f", latest.ContractHours))
			m.inputCommute.SetValue(fmt.Sprintf("%.2f", latest.CommuteHours))
			m.inputErrors.SetValue(fmt.Sprintf("%.2f", latest.ErrorDeductions))
			m.inputCosts.SetValue(fmt.Sprintf("%.2f", latest.InvisibleCosts))
		}
	}

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
		{Title: "ID", Width: 5},
		{Title: "Data", Width: 12},
		{Title: "Bruto (SB)", Width: 14},
		{Title: "VRH", Width: 14},
		{Title: "IDT (%)", Width: 10},
	}

	var rows []table.Row
	if m.repo != nil {
		records, err := m.repo.GetRecentTelemetryRecords(15)
		if err == nil {
			for _, rec := range records {
				dateStr := rec.CreatedAt.Format("2006-01-02")
				rows = append(rows, table.Row{
					fmt.Sprintf("%d", rec.ID),
					dateStr,
					fmt.Sprintf("R$ %.2f", rec.GrossSalary),
					fmt.Sprintf("R$ %.2f/h", rec.VRH),
					fmt.Sprintf("%.2f%%", rec.IDT),
				})
			}
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
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
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Durante a digitacao em um campo (abas com formulario), 'q' e um caractere normal:
			// deixa a tecla cair no roteamento de inputs abaixo. Sem input focado, encerra.
			if m.activeTab != tabHistory {
				break // cai para o routing (que processa o texto no campo focado)
			}
			return m, tea.Quit
		case "1":
			m.activeTab = tabTelemetry
			return m, nil
		case "2":
			m.activeTab = tabBuffer
			return m, nil
		case "3":
			m.activeTab = tabHistory
			return m, nil
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
		if keyMsg, ok := msg.(tea.KeyMsg); ok && (keyMsg.String() == "d" || keyMsg.String() == "delete") {
			selected := m.historyTable.SelectedRow()
			if len(selected) > 0 && m.repo != nil {
				var id int64
				fmt.Sscanf(selected[0], "%d", &id)
				if id > 0 {
					if err := m.repo.DeleteTelemetryRecord(id); err == nil {
						m.statusMsg = fmt.Sprintf("Registro #%d excluído com sucesso!", id)
						m.initHistoryTable()
						return m, nil
					}
				}
			}
		}
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

		if m.repo == nil {
			m.statusMsg = "Telemetria calculada; não foi possível salvar (banco indisponível)."
			return
		}

		refMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
		if _, errSave := m.repo.SaveTelemetry(input, res, refMonth); errSave != nil {
			if strings.Contains(errSave.Error(), "UNIQUE") {
				m.statusMsg = fmt.Sprintf("Telemetria calculada. Registro de competência %s já existia; nada sobrescrito.", refMonth)
			} else {
				m.statusMsg = fmt.Sprintf("Telemetria calculada; não foi possível salvar: %v", errSave)
			}
		} else {
			m.statusMsg = "Telemetria calculada e salva com sucesso!"
		}
		m.initHistoryTable()
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
			refMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
			_, _ = m.repo.SaveBufferCycle(cap, rem, ci, refMonth)
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

	b.WriteString(HelpStyle.Render("[1/2/3] Mudar Aba  |  [Tab/Shift+Tab] Navegar Abas  |  [Seta Cima/Baixo] Campo  |  [Enter] Calcular  |  [q] Sair"))

	return b.String()
}

func (m MainModel) viewTelemetryTab() string {
	var formBuilder strings.Builder
	formBuilder.WriteString(MetricLabelStyle.Render("--- Form Telemetria ---") + "\n")
	formBuilder.WriteString(fmt.Sprintf("Salário Bruto (SB):      %s\n", m.inputSalary.View()))
	formBuilder.WriteString(fmt.Sprintf("Horas Contratuais (HC):  %s\n", m.inputContract.View()))
	formBuilder.WriteString(fmt.Sprintf("Horas Deslocamento (HD): %s\n", m.inputCommute.View()))
	formBuilder.WriteString(fmt.Sprintf("Descontos Erros (DE):    %s\n", m.inputErrors.View()))
	formBuilder.WriteString(fmt.Sprintf("Custos Invisíveis (CI):  %s\n", m.inputCosts.View()))

	formBox := FormBoxStyle.Render(formBuilder.String())

	var resultBox string
	if m.telemetryResult != nil {
		res := m.telemetryResult
		zone, desc := p2t.EvaluateIDTZone(res.IDT)

		var badge string
		var zoneColor = GreenZoneColor
		switch zone {
		case p2t.ZoneGreen:
			badge = BadgeGreenStyle.Render(string(zone))
			zoneColor = GreenZoneColor
		case p2t.ZoneYellow:
			badge = BadgeYellowStyle.Render(string(zone))
			zoneColor = YellowZoneColor
		default:
			badge = BadgeRedStyle.Render(string(zone))
			zoneColor = RedZoneColor
		}

		progressBar := RenderProgressBar(res.IDT*3.33, 24, zoneColor) // 30% max bar scale

		card := fmt.Sprintf(
			"%s\n%s: %.2f h\n%s: %s\n%s: %s\n\n%s: %.2f%%\n[%s]\n\n%s: %s\n%s",
			MetricLabelStyle.Render("--- Resultados & Telemetria ---"),
			MetricLabelStyle.Render("Carga Horária Total (HT)"), res.TotalHours,
			MetricLabelStyle.Render("Liquidez Real (SL)"), MetricMoneyStyle.Render(fmt.Sprintf("R$ %.2f", res.RealLiquidity)),
			MetricLabelStyle.Render("Valor Real Hora (VRH)"), MetricMathStyle.Render(fmt.Sprintf("R$ %.2f/h", res.VRH)),
			MetricLabelStyle.Render("Índice Desperdício (IDT)"), res.IDT,
			progressBar,
			MetricLabelStyle.Render("Matriz de Decisão"), badge,
			desc,
		)
		resultBox = ResultBoxStyle.Render(card)
	} else {
		resultBox = ResultBoxStyle.Render(MetricLabelStyle.Render("Preencha os campos e pressione [Enter] para calcular."))
	}

	if m.width >= 80 {
		return lipgloss.JoinHorizontal(lipgloss.Top, formBox, resultBox)
	}
	return lipgloss.JoinVertical(lipgloss.Left, formBox, resultBox)
}

func (m MainModel) viewBufferTab() string {
	var formBuilder strings.Builder
	formBuilder.WriteString(MetricLabelStyle.Render("--- Gestão do Buffer Operacional ---") + "\n")
	formBuilder.WriteString(fmt.Sprintf("Teto Alocado (T):           %s\n", m.inputCap.View()))
	formBuilder.WriteString(fmt.Sprintf("Saldo Remanescente (S_rem): %s\n", m.inputRemaining.View()))

	formBox := FormBoxStyle.Render(formBuilder.String())

	var resultBox string
	if m.bufferMetrics != nil {
		diag, desc := p2t.DiagnoseEfficiency(m.bufferMetrics.TCM)
		tcmColor := GreenZoneColor
		if m.bufferMetrics.TCM > 50 {
			tcmColor = YellowZoneColor
		}
		if m.bufferMetrics.TCM > 80 {
			tcmColor = RedZoneColor
		}
		bar := RenderProgressBar(m.bufferMetrics.TCM, 24, tcmColor)

		card := fmt.Sprintf(
			"%s\n%s: %s\n%s: %.2f%%\n[%s]\n\n%s: [%s]\n%s",
			MetricLabelStyle.Render("--- Diagnóstico da Caixinha ---"),
			MetricLabelStyle.Render("Média Reposição (R_bar)"), MetricMoneyStyle.Render(fmt.Sprintf("R$ %.2f", m.bufferMetrics.AverageReplenishment)),
			MetricLabelStyle.Render("Taxa Consumo Média (TCM)"), m.bufferMetrics.TCM,
			bar,
			MetricLabelStyle.Render("Status"), diag,
			desc,
		)
		resultBox = ResultBoxStyle.Render(card)
	} else {
		resultBox = ResultBoxStyle.Render(MetricLabelStyle.Render("Preencha Teto e Saldo e pressione [Enter]."))
	}

	if m.width >= 80 {
		return lipgloss.JoinHorizontal(lipgloss.Top, formBox, resultBox)
	}
	return lipgloss.JoinVertical(lipgloss.Left, formBox, resultBox)
}

func (m MainModel) viewHistoryTab() string {
	var b strings.Builder
	b.WriteString(MetricLabelStyle.Render("--- Histórico de Telemetria (SQLite) ---") + "\n\n")
	b.WriteString(m.historyTable.View())
	return b.String()
}
