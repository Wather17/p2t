package cmd

import (
	"fmt"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewGoalCmd cria os subcomandos de gestão de metas financeiras.
func NewGoalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goal",
		Short: "Gerencia metas financeiras sem registrar cada transação",
	}
	cmd.AddCommand(newGoalAddCmd(), newGoalListCmd(), newGoalUpdateCmd(), newGoalArchiveCmd())
	return cmd
}

func openFinanceRepository(dbPath string) (*storage.Repository, func(), error) {
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao abrir banco de dados: %w", err)
	}
	return storage.NewRepository(db), func() { _ = db.Close() }, nil
}

func newGoalAddCmd() *cobra.Command {
	var (
		name                string
		target              float64
		deadline            string
		balance             float64
		monthlyContribution float64
		dbPath              string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Cadastra uma meta financeira",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			id, err := repo.CreateGoal(p2t.Goal{
				Name: name, TargetAmount: target, Deadline: deadline,
				CurrentBalance: balance, MonthlyContribution: monthlyContribution,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Meta #%d criada: %s\n", id, name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Nome da meta")
	cmd.Flags().Float64Var(&target, "target", 0, "Valor alvo da meta")
	cmd.Flags().StringVar(&deadline, "deadline", "", "Prazo da meta (YYYY-MM)")
	cmd.Flags().Float64Var(&balance, "balance", 0, "Saldo atual da meta")
	cmd.Flags().Float64Var(&monthlyContribution, "monthly-contribution", 0, "Aporte mensal planejado")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("target")
	_ = cmd.MarkFlagRequired("deadline")
	return cmd
}

func newGoalListCmd() *cobra.Command {
	var dbPath string
	var includeArchived bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista metas financeiras",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			goals, err := repo.ListGoals(includeArchived)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(goals) == 0 {
				fmt.Fprintln(out, "Nenhuma meta encontrada.")
				return nil
			}
			for _, record := range goals {
				status := "ativa"
				if record.Archived {
					status = "arquivada"
				}
				fmt.Fprintf(out, "#%d %s | alvo R$ %.2f | saldo R$ %.2f (%.2f%%) | prazo %s | aporte R$ %.2f/mês | %s\n",
					record.ID, record.Name, record.TargetAmount, record.CurrentBalance,
					record.ProgressPercent(), record.Deadline, record.MonthlyContribution, status)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&includeArchived, "all", false, "Incluir metas arquivadas")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	return cmd
}

func newGoalUpdateCmd() *cobra.Command {
	var (
		id                  int64
		name                string
		target              float64
		deadline            string
		balance             float64
		monthlyContribution float64
		dbPath              string
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza uma meta financeira",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			goal, err := repo.GetGoal(id)
			if err != nil {
				return err
			}
			if goal == nil {
				return fmt.Errorf("meta #%d nao encontrada", id)
			}
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("target") && !cmd.Flags().Changed("deadline") && !cmd.Flags().Changed("balance") && !cmd.Flags().Changed("monthly-contribution") {
				return fmt.Errorf("informe ao menos um campo para atualizar")
			}
			if cmd.Flags().Changed("name") {
				goal.Name = strings.TrimSpace(name)
			}
			if cmd.Flags().Changed("target") {
				goal.TargetAmount = target
			}
			if cmd.Flags().Changed("deadline") {
				goal.Deadline = deadline
			}
			if cmd.Flags().Changed("balance") {
				goal.CurrentBalance = balance
			}
			if cmd.Flags().Changed("monthly-contribution") {
				goal.MonthlyContribution = monthlyContribution
			}
			if err := repo.UpdateGoal(goal.Goal); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Meta #%d atualizada.\n", id)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "ID da meta")
	cmd.Flags().StringVar(&name, "name", "", "Novo nome da meta")
	cmd.Flags().Float64Var(&target, "target", 0, "Novo valor alvo")
	cmd.Flags().StringVar(&deadline, "deadline", "", "Novo prazo (YYYY-MM)")
	cmd.Flags().Float64Var(&balance, "balance", 0, "Novo saldo atual")
	cmd.Flags().Float64Var(&monthlyContribution, "monthly-contribution", 0, "Novo aporte mensal")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newGoalArchiveCmd() *cobra.Command {
	var id int64
	var dbPath string
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Arquiva uma meta sem apagar seu histórico",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			if err := repo.ArchiveGoal(id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Meta #%d arquivada.\n", id)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "ID da meta")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
