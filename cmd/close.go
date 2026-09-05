package cmd

import (
	"bufio"
	"fmt"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewCloseCmd cria o comando de fechamento financeiro mensal.
func NewCloseCmd() *cobra.Command {
	var (
		referenceMonth    string
		receivedIncome    float64
		reserveBalance    float64
		goalContributions float64
		exceptionalCosts  float64
		dbPath            string
		noSave            bool
		interactive       bool
	)

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Fecha uma competência financeira sem registrar cada transação",
		RunE: func(cmd *cobra.Command, args []string) error {
			in := cmd.InOrStdin()
			out := cmd.OutOrStdout()
			if referenceMonth == "" {
				referenceMonth = time.Now().AddDate(0, -1, 0).Format("2006-01")
			}
			if err := p2t.ValidateReferenceMonth(referenceMonth); err != nil {
				return err
			}

			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()

			goals, err := repo.ListGoals(false)
			if err != nil {
				return err
			}
			commitments, err := repo.GetActiveRecurring()
			if err != nil {
				return err
			}
			var recurringTotal float64
			for _, item := range commitments {
				recurringTotal += item.MonthlyEquivalent()
			}
			recurringTotal = p2t.RoundCurrency(recurringTotal)
			if !cmd.Flags().Changed("goal-contributions") {
				goalContributions = sumGoalContributions(goals)
			}

			if interactive || !cmd.Flags().Changed("received-income") {
				fmt.Fprintf(out, "=== Fechamento Financeiro %s ===\n", referenceMonth)
				scanner := bufio.NewScanner(in)
				if receivedIncome, err = PromptFloat(scanner, out, "Renda recebida no mês", receivedIncome); err != nil {
					return err
				}
				if reserveBalance, err = PromptFloat(scanner, out, "Saldo total das reservas ao final do mês", reserveBalance); err != nil {
					return err
				}
				if goalContributions, err = PromptFloat(scanner, out, "Aportes planejados para metas", sumGoalContributions(goals)); err != nil {
					return err
				}
				if exceptionalCosts, err = PromptFloat(scanner, out, "Custos excepcionais do mês", exceptionalCosts); err != nil {
					return err
				}
			}

			metrics, err := p2t.CalculateFinancialSnapshot(receivedIncome, recurringTotal, goalContributions, exceptionalCosts)
			if err != nil {
				return fmt.Errorf("erro ao calcular fechamento: %w", err)
			}
			snapshot := p2t.FinancialSnapshot{
				ReferenceMonth:           referenceMonth,
				ReceivedIncome:           receivedIncome,
				ReserveBalance:           reserveBalance,
				PlannedGoalContributions: goalContributions,
				ExceptionalCosts:         exceptionalCosts,
				RecurringCommitmentTotal: metrics.RecurringCommitmentTotal,
				GoalCapacity:             metrics.GoalCapacity,
				FreeMargin:               metrics.FreeMargin,
			}

			var telemetry *storage.TelemetryRecord
			telemetry, err = repo.GetTelemetryByReferenceMonth(referenceMonth)
			if err != nil {
				return err
			}

			if !noSave {
				progress := make([]storage.GoalSnapshotProgressRecord, 0, len(goals))
				for _, goal := range goals {
					progress = append(progress, storage.GoalSnapshotProgressRecord{
						GoalID: goal.ID, CurrentBalance: goal.CurrentBalance,
						PlannedContribution: goal.MonthlyContribution,
					})
				}
				if _, err := repo.SaveFinancialSnapshot(snapshot, progress); err != nil {
					return err
				}
				fmt.Fprintf(out, "[Storage] Fechamento salvo para a competência %s.\n", referenceMonth)
			}

			printFinancialSnapshot(out, snapshot, len(goals), len(commitments))
			if telemetry != nil {
				fmt.Fprintf(out, "Telemetria do trabalho: VRH R$ %.2f/h | IDT %.2f%% | horas %.2f | custo direto R$ %.2f\n",
					telemetry.VRH, telemetry.IDT, telemetry.TotalHours,
					p2t.RoundCurrency(telemetry.ErrorDeductions+telemetry.InvisibleCosts))
				printWorkAdvisory(out, snapshot, telemetry)
			} else {
				fmt.Fprintln(out, "Telemetria do trabalho: nenhum registro encontrado para esta competência.")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&referenceMonth, "reference-month", "m", "", "Competência do fechamento (YYYY-MM; padrão: mês anterior)")
	cmd.Flags().Float64Var(&receivedIncome, "received-income", 0, "Renda efetivamente recebida no mês")
	cmd.Flags().Float64Var(&reserveBalance, "reserve-balance", 0, "Saldo total das reservas ao final do mês")
	cmd.Flags().Float64Var(&goalContributions, "goal-contributions", 0, "Aportes planejados para metas")
	cmd.Flags().Float64Var(&exceptionalCosts, "exceptional-costs", 0, "Custos excepcionais do mês")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	cmd.Flags().BoolVar(&noSave, "no-save", false, "Nao salvar o fechamento no banco de dados")
	cmd.Flags().BoolVarP(&interactive, "interactive", "I", false, "Modo guiado por perguntas")
	return cmd
}

func sumGoalContributions(goals []storage.GoalRecord) float64 {
	var total float64
	for _, goal := range goals {
		total += goal.MonthlyContribution
	}
	return p2t.RoundCurrency(total)
}

func printFinancialSnapshot(out interface{ Write([]byte) (int, error) }, snapshot p2t.FinancialSnapshot, goalCount, recurringCount int) {
	fmt.Fprintf(out, "=== Resumo Financeiro %s ===\n", snapshot.ReferenceMonth)
	fmt.Fprintf(out, "Renda Recebida: R$ %.2f\n", snapshot.ReceivedIncome)
	fmt.Fprintf(out, "Compromissos Recorrentes: R$ %.2f (%d itens)\n", snapshot.RecurringCommitmentTotal, recurringCount)
	fmt.Fprintf(out, "Capacidade de Metas: R$ %.2f (%d metas ativas)\n", snapshot.GoalCapacity, goalCount)
	fmt.Fprintf(out, "Aportes Planejados: R$ %.2f\n", snapshot.PlannedGoalContributions)
	fmt.Fprintf(out, "Custos Excepcionais: R$ %.2f\n", snapshot.ExceptionalCosts)
	fmt.Fprintf(out, "Margem Livre: R$ %.2f\n", snapshot.FreeMargin)
	fmt.Fprintf(out, "Reservas: R$ %.2f\n", snapshot.ReserveBalance)
}

func printWorkAdvisory(out interface{ Write([]byte) (int, error) }, snapshot p2t.FinancialSnapshot, telemetry *storage.TelemetryRecord) {
	if snapshot.FreeMargin < 0 {
		fmt.Fprintln(out, "Sinal financeiro: os compromissos, aportes e exceções superaram a renda recebida; revisar o plano.")
	}
	if telemetry.IDT >= p2t.ZoneIDTRedThreshold {
		fmt.Fprintln(out, "Sinal trabalhista: corrosão elevada; acompanhe a tendência e considere um plano de saída.")
	} else if telemetry.IDT >= p2t.ZoneIDTGreenThreshold {
		fmt.Fprintln(out, "Sinal trabalhista: corrosão em observação; não tomar decisão por um mês isolado.")
	}
	if telemetry.VRH <= 0 {
		fmt.Fprintln(out, "Sinal trabalhista: retorno por hora não positivo.")
	}
	if snapshot.FreeMargin >= 0 && telemetry.IDT < p2t.ZoneIDTGreenThreshold && telemetry.VRH > 0 {
		fmt.Fprintln(out, "Sinais advisory: nenhum alerta automático neste fechamento.")
	}
}
