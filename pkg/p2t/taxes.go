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

// CalculateINSS calcula o INSS com base na tabela progressiva oficial.
func CalculateINSS(grossSalary float64) float64 {
	if grossSalary <= 0 {
		return 0.0
	}

	const (
		f1Limit = 1412.00
		f2Limit = 2666.68
		f3Limit = 4000.03
		f4Limit = 7786.02
	)

	var inss float64

	if grossSalary <= f1Limit {
		inss = grossSalary * 0.075
	} else if grossSalary <= f2Limit {
		inss = (f1Limit * 0.075) + ((grossSalary - f1Limit) * 0.09)
	} else if grossSalary <= f3Limit {
		inss = (f1Limit * 0.075) + ((f2Limit - f1Limit) * 0.09) + ((grossSalary - f2Limit) * 0.12)
	} else if grossSalary <= f4Limit {
		inss = (f1Limit * 0.075) + ((f2Limit - f1Limit) * 0.09) + ((f3Limit - f2Limit) * 0.12) + ((grossSalary - f3Limit) * 0.14)
	} else {
		// Teto maximo do INSS
		inss = (f1Limit * 0.075) + ((f2Limit - f1Limit) * 0.09) + ((f3Limit - f2Limit) * 0.12) + ((f4Limit - f3Limit) * 0.14)
	}

	return math.Round(inss*100) / 100
}

// CalculateIRRF calcula o IRRF progressivo a partir do salario bruto e do INSS descontado.
func CalculateIRRF(grossSalary, inss float64) float64 {
	baseCalculation := grossSalary - inss
	if baseCalculation <= 2259.20 {
		return 0.0
	}

	var irrf float64
	switch {
	case baseCalculation <= 2826.65:
		irrf = (baseCalculation * 0.075) - 169.44
	case baseCalculation <= 3751.05:
		irrf = (baseCalculation * 0.15) - 353.44
	case baseCalculation <= 4664.68:
		irrf = (baseCalculation * 0.225) - 634.77
	default:
		irrf = (baseCalculation * 0.275) - 869.36
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
