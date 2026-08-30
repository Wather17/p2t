package p2t

import (
	"math"
)

// TaxBreakdown contem o detalhamento dos descontos legais estimados (INSS + IRRF).
type TaxBreakdown struct {
	INSS            float64
	IRRF            float64
	TotalDeductions float64
}

// CalculateINSS calcula o INSS com base na tabela progressiva oficial (ver tables.go para vigencia).
func CalculateINSS(grossSalary float64) float64 {
	if grossSalary <= 0 {
		return 0.0
	}

	var inss float64
	lowerBound := 0.0
	for _, b := range inssBrackets {
		if grossSalary <= b.limit {
			inss += (grossSalary - lowerBound) * b.rate
			return math.Round(inss*100) / 100
		}
		inss += (b.limit - lowerBound) * b.rate
		lowerBound = b.limit
	}

	// Teto maximo do INSS (salario acima do ultimo limite)
	return math.Round(inss*100) / 100
}

// CalculateIRRF calcula o IRRF progressivo a partir do salario bruto e do INSS descontado.
func CalculateIRRF(grossSalary, inss float64) float64 {
	baseCalculation := grossSalary - inss
	if baseCalculation <= IRRFExemptionLimit {
		return 0.0
	}

	var irrf float64
	for _, b := range irrfBrackets {
		if baseCalculation <= b.limit {
			irrf = (baseCalculation * b.rate) - b.deduction
			break
		}
	}

	if irrf < 0 {
		return 0.0
	}
	return math.Round(irrf*100) / 100
}

// EstimateLegalDeductions calcula automaticamente os descontos legais fixos (INSS + IRRF) para o salario bruto.
func EstimateLegalDeductions(grossSalary float64) TaxBreakdown {
	inss := CalculateINSS(grossSalary)
	irrf := CalculateIRRF(grossSalary, inss)
	total := math.Round((inss+irrf)*100) / 100

	return TaxBreakdown{
		INSS:            inss,
		IRRF:            irrf,
		TotalDeductions: total,
	}
}
