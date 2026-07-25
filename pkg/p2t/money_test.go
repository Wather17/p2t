package p2t_test

import (
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
