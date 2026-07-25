package p2t_test

import (
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
)

func TestCalculateINSS(t *testing.T) {
	tests := []struct {
		salary   float64
		expected float64
	}{
		{1000.0, 75.00},   // 1a faixa: 7.5%
		{2000.0, 158.82},  // 2a faixa
		{3000.0, 258.82},  // 3a faixa
		{5000.0, 518.82},  // 4a faixa
		{10000.0, 908.86}, // Teto INSS
	}

	for _, tt := range tests {
		got := p2t.CalculateINSS(tt.salary)
		if !almostEqual(got, tt.expected) {
			t.Errorf("para salario %.2f esperado INSS %.2f, obtido %.2f", tt.salary, tt.expected, got)
		}
	}
}

func TestEstimateLegalDeductions(t *testing.T) {
	t.Run("Salario Baixo (Isento IRRF)", func(t *testing.T) {
		breakdown := p2t.EstimateLegalDeductions(2000.0)
		if breakdown.INSS <= 0 {
			t.Errorf("esperado INSS > 0, obtido: %.2f", breakdown.INSS)
		}
		if breakdown.IRRF != 0 {
			t.Errorf("esperado IRRF = 0, obtido: %.2f", breakdown.IRRF)
		}
		if !almostEqual(breakdown.TotalDeductions, breakdown.INSS) {
			t.Errorf("esperado TotalDeductions == INSS, obtido: %.2f", breakdown.TotalDeductions)
		}
	})

	t.Run("Salario Medio (Com IRRF)", func(t *testing.T) {
		breakdown := p2t.EstimateLegalDeductions(5000.0)
		if breakdown.INSS <= 0 || breakdown.IRRF <= 0 {
			t.Errorf("esperado INSS e IRRF > 0, obtido INSS=%.2f IRRF=%.2f", breakdown.INSS, breakdown.IRRF)
		}
		if !almostEqual(breakdown.TotalDeductions, breakdown.INSS+breakdown.IRRF) {
			t.Errorf("esperado TotalDeductions == INSS + IRRF, obtido: %.2f", breakdown.TotalDeductions)
		}
	})
}
