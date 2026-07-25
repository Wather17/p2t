package cmd

import (
	"fmt"

	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewHistoryCmd cria o comando history para consultar o historico salvo no SQLite.
func NewHistoryCmd() *cobra.Command {
	var dbPath string
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Exibe o historico de ciclos de telemetria gravados no SQLite",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			db, err := storage.OpenDB(dbPath)
			if err != nil {
				return fmt.Errorf("falha ao abrir banco de dados: %w", err)
			}
			defer db.Close()

			repo := storage.NewRepository(db)
			history, err := repo.GetRecentIDTHistory(limit)
			if err != nil {
				return fmt.Errorf("falha ao obter historico: %w", err)
			}

			fmt.Fprintf(out, "=== Histórico de Telemetria (Últimos %d ciclos) ===\n", limit)
			if len(history) == 0 {
				fmt.Fprintln(out, "Nenhum registro encontrado no banco de dados.")
				return nil
			}

			for i, idt := range history {
				fmt.Fprintf(out, "Ciclo #%d: IDT = %.2f%%\n", i+1, idt)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do banco SQLite (padrao ~/.p2t/p2t.db)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Quantidade de registros a exibir")

	return cmd
}
