package p2t_test

import (
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
)

func TestCalculateInvisibleCost(t *testing.T) {
	t.Run("Calculo Valido", func(t *testing.T) {
		cycle := p2t.BufferCycle{
			Cap:              300.0,
			RemainingBalance: 120.0,
		}
		ci, err := cycle.CalculateInvisibleCost()
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !almostEqual(ci, 180.0) {
			t.Errorf("esperado CI=180.0, obtido=%.2f", ci)
		}
	})

	t.Run("Erro Saldo Maior Que Teto", func(t *testing.T) {
		cycle := p2t.BufferCycle{
			Cap:              300.0,
			RemainingBalance: 350.0,
		}
		_, err := cycle.CalculateInvisibleCost()
		if err == nil {
			t.Error("esperava erro ao tentar saldo remanescente maior que teto")
		}
	})

	t.Run("Erro Teto Invalido", func(t *testing.T) {
		cycle := p2t.BufferCycle{
			Cap:              0,
			RemainingBalance: 50.0,
		}
		_, err := cycle.CalculateInvisibleCost()
		if err == nil {
			t.Error("esperava erro para teto zero")
		}
	})
}

func TestCalculateBufferMetrics(t *testing.T) {
	t.Run("Calculo Valido de TCM e R_bar", func(t *testing.T) {
		cap := 300.0
		replenishments := []float64{100.0, 150.0, 110.0} // media R_bar = 360 / 3 = 120.0
		metrics, err := p2t.CalculateBufferMetrics(cap, replenishments)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if !almostEqual(metrics.AverageReplenishment, 120.0) {
			t.Errorf("esperado R_bar=120.0, obtido=%.2f", metrics.AverageReplenishment)
		}

		// TCM = (120 / 300) * 100 = 40.0%
		if !almostEqual(metrics.TCM, 40.0) {
			t.Errorf("esperado TCM=40.0, obtido=%.2f", metrics.TCM)
		}
	})

	t.Run("Historico Vazio", func(t *testing.T) {
		_, err := p2t.CalculateBufferMetrics(300.0, []float64{})
		if err == nil {
			t.Error("esperava erro para historico vazio")
		}
	})
}

func TestDiagnoseEfficiency(t *testing.T) {
	tests := []struct {
		tcm               float64
		expectedDiagnosis p2t.EfficiencyDiagnosis
	}{
		{40.0, p2t.HighEfficiency},
		{45.0, p2t.StableEfficiency},
		{50.0, p2t.StableEfficiency},
		{55.0, p2t.StableEfficiency},
		{58.0, p2t.AlertEfficiency},
		{61.0, p2t.ConsumptionAnomaly},
	}

	for _, tt := range tests {
		diag, desc := p2t.DiagnoseEfficiency(tt.tcm)
		if diag != tt.expectedDiagnosis {
			t.Errorf("para TCM=%.2f esperado diagnosis %s, obtido %s", tt.tcm, tt.expectedDiagnosis, diag)
		}
		if desc == "" {
			t.Errorf("descricao do diagnostico nao deve ser vazia")
		}
	}
}
