package cmd

import (
	"fmt"

	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewDeleteRecordCmd cria o comando para excluir um registro de telemetria por ID.
func NewDeleteRecordCmd() *cobra.Command {
	var (
		id     int64
		dbPath string
	)

	cmd := &cobra.Command{
		Use:   "delete-record",
		Short: "Exclui um registro de telemetria do banco de dados SQLite por ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if id <= 0 {
				return fmt.Errorf("ID do registro deve ser informado e maior que zero (ex: --id 12)")
			}

			db, err := storage.OpenDB(dbPath)
			if err != nil {
				return fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
			}
			defer db.Close()

			repo := storage.NewRepository(db)
			if err := repo.DeleteTelemetryRecord(id); err != nil {
				return err
			}

			fmt.Fprintf(out, "Registro de telemetria #%d excluído com sucesso do SQLite.\n", id)
			return nil
		},
	}

	cmd.Flags().Int64VarP(&id, "id", "i", 0, "ID do registro de telemetria a ser excluído")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrão ~/.p2t/p2t.db)")

	return cmd
}
