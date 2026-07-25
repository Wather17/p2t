package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// PromptFloat exibe a pergunta e le um float64 usando o scanner fornecido. Se a resposta for vazia e houver defaultValue, usa o defaultValue.
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

		val, err := strconv.ParseFloat(text, 64)
		if err != nil || val < 0 {
			fmt.Fprintln(w, "  -> Entrada inválida. Por favor digite um número maior ou igual a zero.")
			continue
		}

		return val, nil
	}
}
