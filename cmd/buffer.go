package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
	"github.com/Wather17/p2t/pkg/storage"
	"github.com/spf13/cobra"
)

// NewBufferCmd cria o comando de gestao do Buffer Operacional (A Caixinha).
func NewBufferCmd() *cobra.Command {
	var (
		cap               float64
		remainingBalance  float64
		replenishmentsStr string
		dbPath            string
		noSave            bool
		interactive       bool
	)

	cmd := &cobra.Command{
		Use:   "buffer",
		Short: "Gerencia o Buffer Operacional (A Caixinha) e calcula a Taxa de Consumo Media (TCM)",
		RunE: func(cmd *cobra.Command, args []string) error {
			in := cmd.InOrStdin()
			out := cmd.OutOrStdout()

			referenceMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")

			if interactive || cap <= 0 {
				fmt.Fprintln(out, "=== Modo Interativo: Gestão do Buffer Operacional (A Caixinha) ===")
				scanner := bufio.NewScanner(in)
				var err error
				if cap, err = PromptFloat(scanner, out, "Teto fixo alocado para o trabalho (T em R$)", cap); err != nil {
					return err
				}
				if remainingBalance, err = PromptFloat(scanner, out, "Saldo não utilizado retido ao final do ciclo (S_rem em R$)", remainingBalance); err != nil {
					return err
				}
				fmt.Fprintln(out, "================================================================")
			}

			fmt.Fprintf(out, "=== Gestão do Buffer Operacional (A Caixinha) ===\n")

			cycle := p2t.BufferCycle{
				Cap:              cap,
				RemainingBalance: remainingBalance,
			}

			ci, err := cycle.CalculateInvisibleCost()
			if err != nil {
				return fmt.Errorf("erro no calculo do custo invisivel: %w", err)
			}

			fmt.Fprintf(out, "Teto Alocado (T): R$ %.2f\n", cap)
			fmt.Fprintf(out, "Saldo Remanescente (S_rem): R$ %.2f\n", remainingBalance)
			fmt.Fprintf(out, "Custo Invisivel do Ciclo (CI / Rt): R$ %.2f\n", ci)

			var replenishments []float64

			if !noSave {
				db, err := storage.OpenDB(dbPath)
				if err == nil {
					defer db.Close()
					repo := storage.NewRepository(db)

					// Busca historico anterior antes de salvar
					prevRepl, errPrev := repo.GetRecentReplenishments(5)
					if errPrev == nil {
						replenishments = append(replenishments, prevRepl...)
					}

					if _, errSave := repo.SaveBufferCycle(cap, remainingBalance, ci, referenceMonth); errSave == nil {
						fmt.Fprintf(out, "[Storage] Registro de buffer salvo no SQLite com sucesso (competência %s).\n", referenceMonth)
					}
				}
			}

			if replenishmentsStr != "" {
				replenishments = nil
				parts := strings.Split(replenishmentsStr, ",")
				for _, p := range parts {
					v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
					if err != nil {
						return fmt.Errorf("historico de reposicao invalido '%s': %w", p, err)
					}
					replenishments = append(replenishments, v)
				}
			} else if len(replenishments) == 0 {
				replenishments = []float64{ci}
			} else {
				// Inclui o custo invisivel atual no historico
				replenishments = append(replenishments, ci)
			}

			metrics, err := p2t.CalculateBufferMetrics(cap, replenishments)
			if err != nil {
				return fmt.Errorf("erro no calculo de metricas do buffer: %w", err)
			}

			diag, desc := p2t.DiagnoseEfficiency(metrics.TCM)
			fmt.Fprintf(out, "Media Movel de Reposicao (R_bar): R$ %.2f\n", metrics.AverageReplenishment)
			fmt.Fprintf(out, "Taxa de Consumo Media (TCM): %.2f%%\n", metrics.TCM)
			fmt.Fprintf(out, "Diagnostico de Eficiencia: [%s] %s\n", diag, desc)

			return nil
		},
	}

	cmd.Flags().Float64VarP(&cap, "cap", "t", 0, "Teto fixo de liquidez alocado para o trabalho (T)")
	cmd.Flags().Float64VarP(&remainingBalance, "remaining", "r", 0, "Saldo nao utilizado retido ao final do ciclo (S_rem)")
	cmd.Flags().StringVarP(&replenishmentsStr, "replenishments", "p", "", "Historico de valores de reposicao separados por virgula (ex: 120,150,110)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Caminho do arquivo SQLite (padrao ~/.p2t/p2t.db)")
	cmd.Flags().BoolVar(&noSave, "no-save", false, "Nao salvar este registro no banco de dados SQLite")
	cmd.Flags().BoolVarP(&interactive, "interactive", "I", false, "Modo interativo por perguntas")

	return cmd
}
