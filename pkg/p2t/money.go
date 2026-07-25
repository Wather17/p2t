package p2t

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RoundCurrency arredonda um valor monetario para exatamente 2 casas decimais (centavos exatos), evitando imprecisoes do IEEE 754.
func RoundCurrency(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0.0
	}
	return math.Round(val*100.0) / 100.0
}

// RoundPercentage arredonda um valor percentual para 2 casas decimais.
func RoundPercentage(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0.0
	}
	return math.Round(val*100.0) / 100.0
}

// ParseBrazilianFloat converte strings contendo formatos numericos flexiveis/brasileiros (ex: "R$ 5.000,50", "5000,50", "5000.50") em float64.
func ParseBrazilianFloat(input string) (float64, error) {
	str := strings.TrimSpace(input)
	if str == "" {
		return 0, nil
	}

	// Remove prefixos de moeda e simbolos
	str = strings.ReplaceAll(str, "R$", "")
	str = strings.ReplaceAll(str, "$", "")
	str = strings.TrimSpace(str)

	// Lida com pontuacao de milhares e decimais
	lastComma := strings.LastIndex(str, ",")
	lastDot := strings.LastIndex(str, ".")

	if lastComma != -1 && lastDot != -1 {
		if lastComma > lastDot {
			// Formato brasileiro: 5.000,50 -> 5000.50
			str = strings.ReplaceAll(str, ".", "")
			str = strings.ReplaceAll(str, ",", ".")
		} else {
			// Formato americano com separador de milhar: 5,000.50 -> 5000.50
			str = strings.ReplaceAll(str, ",", "")
		}
	} else if lastComma != -1 {
		// Apenas virgula: 5000,50 -> 5000.50
		str = strings.ReplaceAll(str, ",", ".")
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("valor invalido '%s': informe um numero valido", input)
	}

	return RoundCurrency(val), nil
}
