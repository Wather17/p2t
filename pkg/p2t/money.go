package p2t

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// round2 arredonda um valor para exatamente 2 casas decimais, evitando imprecisoes do IEEE 754.
// Implementacao unica: RoundCurrency e RoundPercentage delegaam para ela.
func round2(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0.0
	}
	return math.Round(val*100.0) / 100.0
}

// RoundCurrency arredonda um valor monetario para exatamente 2 casas decimais (centavos exatos), evitando imprecisoes do IEEE 754.
func RoundCurrency(val float64) float64 {
	return round2(val)
}

// RoundPercentage arredonda um valor percentual para 2 casas decimais.
// Equivalente matematico de RoundCurrency (ambos arredondam em 2 casas); mantido apenas por semantica.
func RoundPercentage(val float64) float64 {
	return round2(val)
}

// ParseBrazilianFloat converte strings contendo formatos numericos flexiveis/brasileiros (ex: "R$ 5.000,50", "5000,50", "5000.50") em float64.
//
// Regra de desambiguacao do ponto sem virgula:
//   - "5.000" (ponto unico, 3 digitos apos o ponto e inteiro > 0) -> 5000.00 (ponto e separador de milhar BR);
//   - "1.234.567" -> 1234567.00 (mesma regra, todos os pontos removidos);
//   - "5000.50" (2 digitos) -> 5000.50 (ponto e decimal);
//   - "0.123" (inteiro 0) -> 0.12 (decimal; ambiguidade aceita e documentada: nao interpretamos os 3 digitos como milhar);
//   - "5000.500" -> 5000500.00 (3 digitos com inteiro > 0 = milhar; ambiguidade com precisao decimal de 3 casas aceita).
// Com virgula presente, a virgula sempre vence (formato brasileiro).
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
	} else if lastDot != -1 {
		// Sem virgula: decide entre milhares BR e decimal
		whole := str[:lastDot]
		frac := str[lastDot+1:]
		if len(frac) == 3 && whole != "" && whole != "0" {
			// "5.000" -> milhar BR (remover todos os pontos)
			str = strings.ReplaceAll(str, ".", "")
		}
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("valor invalido '%s': informe um numero valido", input)
	}

	return RoundCurrency(val), nil
}
