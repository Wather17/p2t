package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "0.2.0"
	Commit  = "none"
	Date    = "unknown"
)

// NewVersionCmd cria o comando de versão.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Exibe a versão atual do p2t CLI",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "p2t CLI v%s\n", Version)
			if Commit != "none" {
				fmt.Fprintf(out, " Commit: %s\n", Commit)
			}
			if Date != "unknown" {
				fmt.Fprintf(out, " Build Date: %s\n", Date)
			}
		},
	}
}
