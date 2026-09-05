package cmd

import (
	"fmt"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/spf13/cobra"
)

// NewRecurringCmd cria os subcomandos de gestão de compromissos recorrentes.
func NewRecurringCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recurring",
		Short: "Gerencia assinaturas e outros compromissos recorrentes",
	}
	cmd.AddCommand(newRecurringAddCmd(), newRecurringListCmd(), newRecurringUpdateCmd(), newRecurringArchiveCmd())
	return cmd
}

func newRecurringAddCmd() *cobra.Command {
	var (
		name         string
		amount       float64
		period       string
		purpose      string
		essentiality string
		billingDay   int
		lastReviewed string
		dbPath       string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Cadastra um compromisso recorrente",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			commitment := p2t.RecurringCommitment{
				Name: name, Amount: amount, Period: p2t.RecurringPeriod(period),
				Purpose:      p2t.RecurringPurpose(purpose),
				Essentiality: p2t.RecurringEssentiality(essentiality),
				BillingDay:   billingDay, Status: p2t.RecurringActive, LastReviewed: lastReviewed,
			}
			id, err := repo.CreateRecurring(commitment)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Compromisso #%d criado: %s\n", id, name)
			return nil
		},
	}
	addRecurringFlags(cmd, &name, &amount, &period, &purpose, &essentiality, &billingDay, &lastReviewed, &dbPath)
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("amount")
	_ = cmd.MarkFlagRequired("period")
	_ = cmd.MarkFlagRequired("purpose")
	_ = cmd.MarkFlagRequired("essentiality")
	return cmd
}

func addRecurringFlags(cmd *cobra.Command, name *string, amount *float64, period, purpose, essentiality *string, billingDay *int, lastReviewed, dbPath *string) {
	cmd.Flags().StringVar(name, "name", "", "Nome do compromisso")
	cmd.Flags().Float64Var(amount, "amount", 0, "Valor do compromisso")
	cmd.Flags().StringVar(period, "period", "monthly", "Periodicidade: monthly ou yearly")
	cmd.Flags().StringVar(purpose, "purpose", "life", "Finalidade: life, work ou goal")
	cmd.Flags().StringVar(essentiality, "essentiality", "useful", "Essencialidade: essential, useful ou discretionary")
	cmd.Flags().IntVar(billingDay, "billing-day", 0, "Dia de cobrança (1-31; 0 se não aplicável)")
	cmd.Flags().StringVar(lastReviewed, "last-reviewed", "", "Mês da última revisão (YYYY-MM)")
	cmd.Flags().StringVar(dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
}

func newRecurringListCmd() *cobra.Command {
	var dbPath string
	var includeArchived bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista compromissos recorrentes",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			items, err := repo.ListRecurring(includeArchived)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(items) == 0 {
				fmt.Fprintln(out, "Nenhum compromisso recorrente encontrado.")
				return nil
			}
			for _, item := range items {
				status := "ativo"
				if item.Status == p2t.RecurringArchived {
					status = "arquivado"
				}
				fmt.Fprintf(out, "#%d %s | R$ %.2f/%s (R$ %.2f/mês) | %s/%s | cobrança dia %d | revisão %s | %s\n",
					item.ID, item.Name, item.Amount, item.Period, item.MonthlyEquivalent(),
					item.Purpose, item.Essentiality, item.BillingDay, emptyAsDash(item.LastReviewed), status)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&includeArchived, "all", false, "Incluir compromissos arquivados")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	return cmd
}

func newRecurringUpdateCmd() *cobra.Command {
	var (
		id           int64
		name         string
		amount       float64
		period       string
		purpose      string
		essentiality string
		billingDay   int
		lastReviewed string
		dbPath       string
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza um compromisso recorrente",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			record, err := repo.GetRecurring(id)
			if err != nil {
				return err
			}
			if record == nil {
				return fmt.Errorf("compromisso #%d nao encontrado", id)
			}
			changed := false
			if cmd.Flags().Changed("name") {
				record.Name = strings.TrimSpace(name)
				changed = true
			}
			if cmd.Flags().Changed("amount") {
				record.Amount = amount
				changed = true
			}
			if cmd.Flags().Changed("period") {
				record.Period = p2t.RecurringPeriod(period)
				changed = true
			}
			if cmd.Flags().Changed("purpose") {
				record.Purpose = p2t.RecurringPurpose(purpose)
				changed = true
			}
			if cmd.Flags().Changed("essentiality") {
				record.Essentiality = p2t.RecurringEssentiality(essentiality)
				changed = true
			}
			if cmd.Flags().Changed("billing-day") {
				record.BillingDay = billingDay
				changed = true
			}
			if cmd.Flags().Changed("last-reviewed") {
				record.LastReviewed = lastReviewed
				changed = true
			}
			if !changed {
				return fmt.Errorf("informe ao menos um campo para atualizar")
			}
			if err := repo.UpdateRecurring(record.RecurringCommitment); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Compromisso #%d atualizado.\n", id)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "ID do compromisso")
	cmd.Flags().StringVar(&name, "name", "", "Novo nome")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Novo valor")
	cmd.Flags().StringVar(&period, "period", "", "Nova periodicidade: monthly ou yearly")
	cmd.Flags().StringVar(&purpose, "purpose", "", "Nova finalidade: life, work ou goal")
	cmd.Flags().StringVar(&essentiality, "essentiality", "", "Nova essencialidade: essential, useful ou discretionary")
	cmd.Flags().IntVar(&billingDay, "billing-day", 0, "Novo dia de cobrança")
	cmd.Flags().StringVar(&lastReviewed, "last-reviewed", "", "Novo mês de revisão (YYYY-MM)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newRecurringArchiveCmd() *cobra.Command {
	var id int64
	var dbPath string
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Arquiva um compromisso sem apagar seu histórico",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closeDB, err := openFinanceRepository(dbPath)
			if err != nil {
				return err
			}
			defer closeDB()
			if err := repo.ArchiveRecurring(id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Compromisso #%d arquivado.\n", id)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "ID do compromisso")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
