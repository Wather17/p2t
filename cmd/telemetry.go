package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewTelemetryCmd cria o comando de telemetria de retorno de tempo (VRH e IDT).
func NewTelemetryCmd() *cobra.Command {
	var (
		grossSalary     float64
		fixedDeductions float64
		errorDeductions float64
		invisibleCosts  float64
		contractHours   float64
		commuteHours    float64
		idtHistoryStr   string
		dbPath          string
		noSave          bool
		interactive     bool
	)

	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Calcula o VRH (Valor Real da Hora Dedicada) e o IDT (Indice de Desperdicio de Trabalho)",
		RunE: func(cmd *cobra.Command, args []string) error {
			in := cmd.InOrStdin()
			out := cmd.OutOrStdout()

			// Se a flag interactive for explicitamente usada ou se os valores obrigatorios forem omissos
			if interactive || (grossSalary <= 0 && contractHours <= 0) {
				fmt.Fprintln(out, "=== Modo Interativo: Telemetria p2t ===")
				scanner := bufio.NewScanner(in)
				var err error

				if grossSalary, err = PromptFloat(scanner, out, "Salário Bruto Nominal (SB em R$)", grossSalary); err != nil {
					return err
				}
				if contractHours, err = PromptFloat(scanner, out, "Carga Horária Mensal Contratual (HC em horas)", contractHours); err != nil {
					return err
				}
				if commuteHours, err = PromptFloat(scanner, out, "Horas Mensais gastas exclusivamente em Deslocamento (HD)", commuteHours); err != nil {
					return err
				}

				// Calcula estimativa automatica de descontos legais se nao especificado
				estimatedTaxes := p2t.EstimateLegalDeductions(grossSalary)
				defaultDF := fixedDeductions
				if defaultDF <= 0 {
					defaultDF = estimatedTaxes.TotalDeductions
				}

				promptDFMsg := fmt.Sprintf("Descontos Legais Fixos (DF - INSS: R$ %.2f, IRRF: R$ %.2f)", estimatedTaxes.INSS, estimatedTaxes.IRRF)
				if fixedDeductions, err = PromptFloat(scanner, out, promptDFMsg, defaultDF); err != nil {
					return err
				}
				if errorDeductions, err = PromptFloat(scanner, out, "Descontos por Erros Operacionais (DE)", errorDeductions); err != nil {
					return err
				}
				if invisibleCosts, err = PromptFloat(scanner, out, "Custos Invisíveis do Bolso para trabalhar (CI)", invisibleCosts); err != nil {
					return err
				}
				fmt.Fprintln(out, "========================================")
			} else if fixedDeductions <= 0 && grossSalary > 0 {
				// Se nao forneceu os descontos fixos nas flags, calcula a estimativa automatica oficial
				estimatedTaxes := p2t.EstimateLegalDeductions(grossSalary)
				fixedDeductions = estimatedTaxes.TotalDeductions
			}

			input := p2t.TelemetryInput{
				GrossSalary:     grossSalary,
				FixedDeductions: fixedDeductions,
				ErrorDeductions: errorDeductions,
				InvisibleCosts:  invisibleCosts,
				ContractHours:   contractHours,
				CommuteHours:    commuteHours,
			}

			res, err := p2t.CalculateTelemetry(input)
			if err != nil {
				return fmt.Errorf("erro no calculo de telemetria: %w", err)
			}

			fmt.Fprintf(out, "=== Telemetria de Retorno de Tempo (p2t) ===\n")
			fmt.Fprintf(out, "Carga Horaria Total (HT): %.2f h\n", res.TotalHours)
			fmt.Fprintf(out, "Descontos Legais Aplicados (DF): R$ %.2f\n", fixedDeductions)
			fmt.Fprintf(out, "Liquidez Real (SL): R$ %.2f\n", res.RealLiquidity)
			fmt.Fprintf(out, "Valor Real da Hora (VRH): R$ %.2f / h\n", res.VRH)
			fmt.Fprintf(out, "Indice de Desperdicio (IDT): %.2f%%\n", res.IDT)

			var history []float64

			if !noSave {
				db, err := storage.OpenDB(dbPath)
				if err == nil {
					defer db.Close()
					repo := storage.NewRepository(db)

					// Busca historico anterior antes de salvar o ciclo atual
					prevHistory, errPrev := repo.GetRecentIDTHistory(2)
					if errPrev == nil {
						history = append(history, prevHistory...)
					}

					if _, errSave := repo.SaveTelemetry(input, res); errSave == nil {
						fmt.Fprintf(out, "[Storage] Ciclo registrado no SQLite com sucesso.\n")
					}
				}
			}

			if idtHistoryStr != "" {
				parts := strings.Split(idtHistoryStr, ",")
				history = nil
				for _, p := range parts {
					v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
					if err != nil {
						return fmt.Errorf("historico IDT invalido '%s': %w", p, err)
					}
					history = append(history, v)
				}
			}

			// Inclui o IDT atual no historico para calculo da media movel
			history = append(history, res.IDT)

			idt3, err := p2t.CalculateIDT3(history)
			if err != nil {
				zone, desc := p2t.EvaluateIDTZone(res.IDT)
				fmt.Fprintf(out, "Matriz de Decisao (Ciclo Atual): [%s] %s\n", zone, desc)
			} else {
				zone, desc := p2t.EvaluateIDTZone(idt3)
				fmt.Fprintf(out, "Média Móvel (IDT3): %.2f%%\n", idt3)
				fmt.Fprintf(out, "Matriz de Decisao: [%s] %s\n", zone, desc)
			}

			return nil
		},
	}

	cmd.Flags().Float64VarP(&grossSalary, "gross-salary", "s", 0, "Salario Bruto (SB)")
	cmd.Flags().Float64VarP(&fixedDeductions, "fixed-deductions", "f", 0, "Descontos Legais Fixos (DF). Se 0, calcula INSS+IRRF automaticamente.")
	cmd.Flags().Float64VarP(&errorDeductions, "error-deductions", "e", 0, "Descontos por Erros Operacionais (DE)")
	cmd.Flags().Float64VarP(&invisibleCosts, "invisible-costs", "c", 0, "Custos Invisiveis de Permanencia (CI)")
	cmd.Flags().Float64VarP(&contractHours, "contract-hours", "H", 0, "Horas Mensais Contratuais (HC)")
	cmd.Flags().Float64VarP(&commuteHours, "commute-hours", "d", 0, "Horas Mensais de Deslocamento (HD)")
	cmd.Flags().StringVarP(&idtHistoryStr, "idt-history", "i", "", "Historico de IDTs anteriores separados por virgula (ex: 8.5,9.0)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	cmd.Flags().BoolVar(&noSave, "no-save", false, "Nao salvar este registro no banco de dados SQLite")
	cmd.Flags().BoolVarP(&interactive, "interactive", "I", false, "Modo interativo por perguntas")

	return cmd
}
