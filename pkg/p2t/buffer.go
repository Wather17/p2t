package p2t

import (
	"errors"
	"fmt"
)

// EfficiencyDiagnosis representa o resultado da auditoria de consumo do buffer operacional.
type EfficiencyDiagnosis string

const (
	HighEfficiency     EfficiencyDiagnosis = "Alta Eficiencia"
	StableEfficiency   EfficiencyDiagnosis = "Eficiencia Estavel"
	ConsumptionAnomaly EfficiencyDiagnosis = "Anomalia de Consumo"
	AlertEfficiency     EfficiencyDiagnosis = "Alerta de Consumo Elevado"
)

// BufferCycle representa o registro de um ciclo mensal da Caixinha.
type BufferCycle struct {
	Cap              float64 // T: Teto fixo alocado
	RemainingBalance float64 // S_rem: Saldo nao utilizado ao final do ciclo
}

// CalculateInvisibleCost calcula o C_I (custo invisivel/reposicao) do ciclo: C_I = T - S_rem com precisao monetaria.
func (b BufferCycle) CalculateInvisibleCost() (float64, error) {
	if b.Cap <= 0 {
		return 0, errors.New("teto do buffer (T) deve ser maior que zero")
	}
	if b.RemainingBalance < 0 {
		return 0, errors.New("saldo remanescente (S_rem) nao pode ser negativo")
	}
	if b.RemainingBalance > b.Cap {
		return 0, fmt.Errorf("saldo remanescente (%.2f) nao pode exceder o teto (%.2f)", b.RemainingBalance, b.Cap)
	}
	return RoundCurrency(b.Cap - b.RemainingBalance), nil
}

// BufferMetrics guarda os resultados agregados de reposicao e consumo do buffer.
type BufferMetrics struct {
	AverageReplenishment float64 // R_bar: Media movel de reposicao
	TCM                  float64 // TCM: Taxa de Consumo Media (%)
}

// CalculateBufferMetrics calcula a media movel de reposicao R_bar e o TCM dado um historico de reposicoes e o teto T.
func CalculateBufferMetrics(cap float64, replenishments []float64) (BufferMetrics, error) {
	if cap <= 0 {
		return BufferMetrics{}, errors.New("teto do buffer (T) deve ser maior que zero")
	}
	if len(replenishments) == 0 {
		return BufferMetrics{}, errors.New("historico de reposicoes nao pode ser vazio")
	}

	var total float64
	for i, r := range replenishments {
		if r < 0 {
			return BufferMetrics{}, fmt.Errorf("valor de reposicao invalido no indice %d: %.2f", i, r)
		}
		total += r
	}

	rBar := RoundCurrency(total / float64(len(replenishments)))
	tcm := RoundPercentage((rBar / cap) * 100.0)

	return BufferMetrics{
		AverageReplenishment: rBar,
		TCM:                  tcm,
	}, nil
}

// DiagnoseEfficiency classifica a taxa de consumo media (TCM).
func DiagnoseEfficiency(tcm float64) (EfficiencyDiagnosis, string) {
	tcm = RoundPercentage(tcm)
	switch {
	case tcm < ZoneTCMHighThreshold:
		return HighEfficiency, "Otimizacao de custos de transporte/alimentacao"
	case tcm >= ZoneTCMHighThreshold && tcm <= ZoneTCMAlertThreshold:
		return StableEfficiency, "Alinhada às estimativas de projeto"
	case tcm > ZoneTCMAlertThreshold && tcm <= ZoneTCMAnomalyThreshold:
		return AlertEfficiency, "Atencao: Consumo proximo do limite critico"
	default:
		return ConsumptionAnomaly, "Requer auditoria pontual do extrato bancario"
	}
}
