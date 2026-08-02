package cmd

import (
	"fmt"
	"strings"

	"github.com/Wather17/p2t/pkg/tui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// NewMathCmd cria o comando de documentacao e fundamentacao matematica do p2t.
func NewMathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "math",
		Aliases: []string{"docs", "specification"},
		Short:   "Exibe a fundamentação matemática, glossário de incógnitas e equações dedutivas do p2t",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			var b strings.Builder

			// Banner Principal
			b.WriteString(tui.AppTitleStyle.Render("p2t - Fundamentação Matemática & Método Dedutivo"))
			b.WriteString("\n\n")

			// Secao 1: Glossario de Incognitas
			section1 := fmt.Sprintf(
				"%s\n"+
					"%s: Salário Bruto nominal registrado em contrato.\n"+
					"%s: Descontos Legais Fixos (INSS, IRRF, impostos de folha).\n"+
					"%s: Descontos decorrentes de erros operacionais / penalidades.\n"+
					"%s: Custos Invisíveis do bolso para trabalhar (transporte, alimentação extra).\n"+
					"%s: Horas mensais contratualmente devidas.\n"+
					"%s: Horas mensais gastas exclusivamente em deslocamento (ida/volta).\n"+
					"%s: Carga Horária Total Efetivamente Dedicada (HT = HC + HD).\n"+
					"%s: Liquidez Real Disponível (SL = SB - (DF + DE + CI)).\n",
				tui.MetricLabelStyle.Render("=== 1. GLOSSÁRIO DE INCÓGNITAS ==="),
				tui.MetricMathStyle.Render("SB"),
				tui.MetricMathStyle.Render("DF"),
				tui.MetricMathStyle.Render("DE"),
				tui.MetricMathStyle.Render("CI"),
				tui.MetricMathStyle.Render("HC"),
				tui.MetricMathStyle.Render("HD"),
				tui.MetricMathStyle.Render("HT"),
				tui.MetricMathStyle.Render("SL"),
			)
			b.WriteString(tui.FormBoxStyle.Render(section1))
			b.WriteString("\n\n")

			// Secao 2: Equacoes Principais
			section2 := fmt.Sprintf(
				"%s\n\n"+
					"%s (Valor Real da Hora Dedicada):\n"+
					"  VRH = [ SB - (DF + DE + CI) ] / (HC + HD)\n"+
					"  Retorno financeiro por hora de vida entregue ao emprego.\n\n"+
					"%s (Índice de Desperdício de Trabalho):\n"+
					"  IDT = [ (DE + CI) / SB ] * 100\n"+
					"  Porcentagem do salário bruto corroída por custos operacionais de permanência.\n\n"+
					"%s (Taxa de Consumo Média do Buffer/Caixinha):\n"+
					"  TCM = (R_bar / T) * 100\n"+
					"  Proporção de utilização do teto (T) pela média de reposições (R_bar).\n",
				tui.MetricLabelStyle.Render("=== 2. EQUAÇÕES FUNDAMENTAIS ==="),
				tui.MetricMathStyle.Render("VRH"),
				tui.MetricMathStyle.Render("IDT"),
				tui.MetricMathStyle.Render("TCM"),
			)
			b.WriteString(tui.FormBoxStyle.Render(section2))
			b.WriteString("\n\n")

			// Secao 3: Matriz de Decisao
			section3 := fmt.Sprintf(
				"%s\n\n"+
					"%s  IDT3 < 10%%   : Estável / Padrão Operacional Mantido.\n"+
					"%s IDT3 10-15%% : Alerta de Corrosão / Ativar Busca Passiva de Vagas.\n"+
					"%s    IDT3 >= 15%%  : Inviabilidade Financeira / Gatilho de Saída Ativado.\n",
				tui.MetricLabelStyle.Render("=== 3. MATRIZ DE DECISÃO (MÉDIA MÓVEL IDT3) ==="),
				tui.BadgeGreenStyle.Render("ZONA VERDE"),
				tui.BadgeYellowStyle.Render("ZONA AMARELA"),
				tui.BadgeRedStyle.Render("ZONA VERMELHA"),
			)
			b.WriteString(tui.ResultBoxStyle.Render(section3))
			b.WriteString("\n\n")

			// Secao 4: Exemplo Resolvido
			exampleBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(tui.GreenSecondary).
				Padding(1, 2).
				Render(
					fmt.Sprintf(
						"%s\n"+
							"Dados: SB = R$ 5.000,00 | DF = R$ 800,00 | CI = R$ 200,00 | DE = R$ 0,00\n"+
							"       HC = 160h | HD = 33h (Escala 5x2 com 1.5h/dia trânsito)\n\n"+
							"Cálculo:\n"+
							"  HT = 160 + 33 = 193h\n"+
							"  SL = 5000 - (800 + 200) = R$ 4.000,00\n"+
							"  VRH = 4000 / 193 = %s\n"+
							"  IDT = (200 / 5000) * 100 = %s -> %s\n",
						tui.MetricLabelStyle.Render("=== 4. EXEMPLO PRÁTICO RESOLVIDO ==="),
						tui.MetricMathStyle.Render("R$ 20.73 / h"),
						tui.MetricMoneyStyle.Render("4.00%"),
						tui.BadgeGreenStyle.Render("Zona Verde"),
					),
				)
			b.WriteString(exampleBox)
			b.WriteString("\n")

			fmt.Fprintln(out, b.String())
			return nil
		},
	}

	return cmd
}
