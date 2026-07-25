package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
)

// PromptFloat exibe a pergunta e le um float64 aceitando formatos brasileiros (virgula, R$, milhares).
func PromptFloat(scanner *bufio.Scanner, w io.Writer, question string, defaultValue float64) (float64, error) {
	for {
		if defaultValue > 0 {
			fmt.Fprintf(w, "? %s [Padrão: %.2f]: ", question, defaultValue)
		} else {
			fmt.Fprintf(w, "? %s: ", question)
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return 0, err
			}
			return defaultValue, nil
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			if defaultValue >= 0 {
				return defaultValue, nil
			}
			fmt.Fprintln(w, "  -> Campo obrigatório. Por favor informe um valor válido.")
			continue
		}

		val, err := p2t.ParseBrazilianFloat(text)
		if err != nil || val < 0 {
			fmt.Fprintln(w, "  -> Entrada inválida. Digite um número positivo (ex: 5000,50 ou R$ 5.000,50).")
			continue
		}

		return val, nil
	}
}
