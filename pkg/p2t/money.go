package p2t

import (
	"math"
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
