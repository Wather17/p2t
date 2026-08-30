package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewImportCmd cria o comando para importar registros de histórico a partir de um arquivo JSON ou CSV.
func NewImportCmd() *cobra.Command {
	var (
		file   string
		dbPath string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Importa registros de histórico a partir de um arquivo JSON ou CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if file == "" {
				return fmt.Errorf("caminho do arquivo para importação é obrigatório (use --file)")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("falha ao ler arquivo de importação '%s': %w", file, err)
			}

			db, err := storage.OpenDB(dbPath)
			if err != nil {
				return fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
			}
			defer db.Close()

			repo := storage.NewRepository(db)
			ext := strings.ToLower(filepath.Ext(file))

			var result storage.ImportResult
			if ext == ".csv" || strings.HasPrefix(strings.TrimSpace(string(data)), "id,") {
				result, err = storage.ImportTelemetryCSV(repo, data)
			} else {
				result, err = storage.ImportTelemetryJSON(repo, data)
			}

			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Importação concluída: %d registro(s) inserido(s) no SQLite.\n", result.Inserted)
			if result.Duplicated > 0 {
				fmt.Fprintf(out, "Resumo: %d duplicado(s) ignorado(s) (competência já existente), %d falha(s).\n", result.Duplicated, result.Failed)
			} else {
				fmt.Fprintf(out, "Resumo: %d falha(s).\n", result.Failed)
			}
			if result.Duplicated > 0 || result.Failed > 0 {
				for _, reason := range result.Why {
					fmt.Fprintf(out, "  - %s\n", reason)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "F", "", "Caminho do arquivo de importação (JSON ou CSV)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrão ~/.p2t/p2t.db)")

	return cmd
}
