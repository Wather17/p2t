package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "p2t",
	Short: "p2t (Pay to Work) - Telemetria de Eficiência de Trabalho",
	Long: `p2t é uma ferramenta CLI/TUI para mensurar o retorno real do tempo investido no trabalho (VRH e IDT) 
e gerenciar por exceção os custos invisíveis com o Buffer Operacional (A Caixinha).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Se executado apenas como 'p2t', abre a TUI por padrao
		tuiCmd := NewTUICmd()
		return tuiCmd.RunE(tuiCmd, args)
	},
}

// Execute executa o comando raiz da CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NewRootCmd cria e retorna uma nova instancia do rootCmd para testes e execucao programatica.
func NewRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "p2t",
		Short: "p2t (Pay to Work) - Telemetria de Eficiência de Trabalho",
		Long: `p2t é uma ferramenta CLI/TUI para mensurar o retorno real do tempo investido no trabalho (VRH e IDT) 
e gerenciar por exceção os custos invisíveis com o Buffer Operacional (A Caixinha).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tuiCmd := NewTUICmd()
			return tuiCmd.RunE(tuiCmd, args)
		},
	}
	c.AddCommand(NewTUICmd())
	c.AddCommand(NewTelemetryCmd())
	c.AddCommand(NewBufferCmd())
	c.AddCommand(NewHistoryCmd())
	c.AddCommand(NewVersionCmd())
	c.AddCommand(NewMathCmd())
	c.AddCommand(NewDeleteRecordCmd())
	return c
}

func init() {
	rootCmd.AddCommand(NewTUICmd())
	rootCmd.AddCommand(NewTelemetryCmd())
	rootCmd.AddCommand(NewBufferCmd())
	rootCmd.AddCommand(NewHistoryCmd())
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewMathCmd())
	rootCmd.AddCommand(NewDeleteRecordCmd())
}
