package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewExportCmd cria o comando para exportar os dados do SQLite em formato JSON ou CSV.
func NewExportCmd() *cobra.Command {
	var (
		format string
		output string
		dbPath string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exporta os registros de histórico do p2t para JSON ou CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			db, err := storage.OpenDB(dbPath)
			if err != nil {
				return fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
			}
			defer db.Close()

			repo := storage.NewRepository(db)

			var payload []byte
			format = strings.ToLower(format)

			switch format {
			case "json":
				payload, err = storage.ExportTelemetryJSON(repo)
			case "csv":
				payload, err = storage.ExportTelemetryCSV(repo)
			default:
				return fmt.Errorf("formato não suportado '%s': utilize 'json' ou 'csv'", format)
			}

			if err != nil {
				return err
			}

			if output == "" || output == "-" {
				fmt.Fprintln(out, string(payload))
			} else {
				if err := os.WriteFile(output, payload, 0644); err != nil {
					return fmt.Errorf("falha ao escrever no arquivo '%s': %w", output, err)
				}
				fmt.Fprintf(out, "Histórico exportado com sucesso para %s (%s, %d bytes).\n", output, strings.ToUpper(format), len(payload))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "json", "Formato de exportação ('json' ou 'csv')")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Caminho do arquivo de destino (saída padrão se omitido)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrão ~/.p2t/p2t.db)")

	return cmd
}
