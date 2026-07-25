package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
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
	)

	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Calcula o VRH (Valor Real da Hora Dedicada) e o IDT (Indice de Desperdicio de Trabalho)",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "=== Telemetria de Retorno de Tempo (p2t) ===\n")
			fmt.Fprintf(out, "Carga Horaria Total (HT): %.2f h\n", res.TotalHours)
			fmt.Fprintf(out, "Liquidez Real (SL): R$ %.2f\n", res.RealLiquidity)
			fmt.Fprintf(out, "Valor Real da Hora (VRH): R$ %.2f / h\n", res.VRH)
			fmt.Fprintf(out, "Indice de Desperdicio (IDT): %.2f%%\n", res.IDT)

			if idtHistoryStr != "" {
				parts := strings.Split(idtHistoryStr, ",")
				var history []float64
				for _, p := range parts {
					v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
					if err != nil {
						return fmt.Errorf("historico IDT invalido '%s': %w", p, err)
					}
					history = append(history, v)
				}
				// Inclui o IDT atual no final do historico se solicitado
				history = append(history, res.IDT)
				idt3, err := p2t.CalculateIDT3(history)
				if err != nil {
					fmt.Fprintf(out, "Média Móvel (IDT3): Historico insuficiente (%v)\n", err)
				} else {
					zone, desc := p2t.EvaluateIDTZone(idt3)
					fmt.Fprintf(out, "Média Móvel (IDT3): %.2f%%\n", idt3)
					fmt.Fprintf(out, "Matriz de Decisao: [%s] %s\n", zone, desc)
				}
			} else {
				zone, desc := p2t.EvaluateIDTZone(res.IDT)
				fmt.Fprintf(out, "Matriz de Decisao (Ciclo Atual): [%s] %s\n", zone, desc)
			}

			return nil
		},
	}

	cmd.Flags().Float64VarP(&grossSalary, "gross-salary", "s", 0, "Salario Bruto (SB)")
	cmd.Flags().Float64VarP(&fixedDeductions, "fixed-deductions", "f", 0, "Descontos Legais Fixos (DF)")
	cmd.Flags().Float64VarP(&errorDeductions, "error-deductions", "e", 0, "Descontos por Erros Operacionais (DE)")
	cmd.Flags().Float64VarP(&invisibleCosts, "invisible-costs", "c", 0, "Custos Invisiveis de Permanencia (CI)")
	cmd.Flags().Float64VarP(&contractHours, "contract-hours", "H", 0, "Horas Mensais Contratuais (HC)")
	cmd.Flags().Float64VarP(&commuteHours, "commute-hours", "d", 0, "Horas Mensais de Deslocamento (HD)")
	cmd.Flags().StringVarP(&idtHistoryStr, "idt-history", "i", "", "Historico de IDTs anteriores separados por virgula (ex: 8.5,9.0)")

	_ = cmd.MarkFlagRequired("gross-salary")
	_ = cmd.MarkFlagRequired("contract-hours")

	return cmd
}
