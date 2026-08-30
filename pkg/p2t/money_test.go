package p2t_test

import (
	"math"
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
)

func TestMoneyRounding_IEEE754(t *testing.T) {
	t.Run("RoundCurrency com dizimas e floating point noise", func(t *testing.T) {
		val := 0.1 + 0.2 // no IEEE 754 float puro resulta em 0.30000000000000004
		got := p2t.RoundCurrency(val)
		if got != 0.30 {
			t.Errorf("esperado 0.30, obtido %.17f", got)
		}
	})

	t.Run("RoundCurrency com centavos fracionados", func(t *testing.T) {
		tests := []struct {
			input    float64
			expected float64
		}{
			{10.556, 10.56},
			{10.554, 10.55},
			{18.499999, 18.50},
		}

		for _, tt := range tests {
			got := p2t.RoundCurrency(tt.input)
			if got != tt.expected {
				t.Errorf("para input %.6f esperado %.2f, obtido %.2f", tt.input, tt.expected, got)
			}
		}
	})
}

func TestParseBrazilianFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5000", 5000.00},
		{"5000,50", 5000.50},
		{"5000.50", 5000.50},
		{"5.000,50", 5000.50},
		{"R$ 5.000,50", 5000.50},
		{"R$5000,50", 5000.50},
		{"$ 1.250.000,75", 1250000.75},
		{"160,0", 160.00},
		{"5.000", 5000.00},
		{"R$ 5.000", 5000.00},
		{"1.234.567", 1234567.00},
		{"10.500", 10500.00},
		{"0.123", 0.12},
	}

	for _, tt := range tests {
		got, err := p2t.ParseBrazilianFloat(tt.input)
		if err != nil {
			t.Errorf("erro inesperado para input %s: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("para input %s esperado %.2f, obtido %.2f", tt.input, tt.expected, got)
		}
	}
}

func TestParseBrazilianFloat_Errors(t *testing.T) {
	inputs := []string{"abc", "1,2,3", "R$ 90,xx"}

	for _, input := range inputs {
		if _, err := p2t.ParseBrazilianFloat(input); err == nil {
			t.Errorf("esperado erro para input %s, obtido nil", input)
		}
	}
}

func TestRoundingEquivalence(t *testing.T) {
	inputs := []float64{20.456, 0.1 + 0.2, 1234.565, 15.5, 7.25, 0.0, -3.1415}

	for _, v := range inputs {
		if got, want := p2t.RoundCurrency(v), p2t.RoundPercentage(v); got != want {
			t.Errorf("para input %v: RoundCurrency=%v != RoundPercentage=%v", v, got, want)
		}
	}

	for _, v := range []float64{math.NaN(), math.Inf(1)} {
		if got := p2t.RoundCurrency(v); got != 0.0 {
			t.Errorf("esperado 0.0 para valor invalido, obtido %v", got)
		}
		if got := p2t.RoundPercentage(v); got != 0.0 {
			t.Errorf("esperado 0.0 para valor invalido, obtido %v", got)
		}
	}
}
