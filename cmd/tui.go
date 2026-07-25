package cmd

import (
	"fmt"

	"github.com/Wather17/p2t/pkg/storage"
	"github.com/Wather17/p2t/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// NewTUICmd cria o comando tui para abrir a interface gráfica de terminal.
func NewTUICmd() *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Abre a TUI (Terminal User Interface) interativa do p2t",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := storage.OpenDB(dbPath)
			if err != nil {
				return fmt.Errorf("falha ao inicializar banco de dados para TUI: %w", err)
			}
			defer db.Close()

			model := tui.NewMainModel(db)
			p := tea.NewProgram(model, tea.WithAltScreen())

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("erro na execucao da TUI: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do banco de dados SQLite (padrao ~/.p2t/p2t.db)")
	return cmd
}
