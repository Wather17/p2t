package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "p2t",
	Short: "p2t (Pay to Work) - Telemetria de Eficiencia de Trabalho",
	Long: `p2t e uma ferramenta CLI/TUI para mensurar o retorno real do tempo investido no trabalho (VRH e IDT) 
e gerenciar por excecao os custos invisiveis com o Buffer Operacional (A Caixinha).`,
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
		Short: "p2t (Pay to Work) - Telemetria de Eficiencia de Trabalho",
		Long: `p2t e uma ferramenta CLI/TUI para mensurar o retorno real do tempo investido no trabalho (VRH e IDT) 
e gerenciar por excecao os custos invisiveis com o Buffer Operacional (A Caixinha).`,
	}
	c.AddCommand(NewTelemetryCmd())
	c.AddCommand(NewBufferCmd())
	c.AddCommand(NewVersionCmd())
	return c
}

func init() {
	rootCmd.AddCommand(NewTelemetryCmd())
	rootCmd.AddCommand(NewBufferCmd())
	rootCmd.AddCommand(NewVersionCmd())
}
