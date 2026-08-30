package p2t

import "math"

// Este arquivo centraliza as tabelas e limiares do dominio p2t para facilitar a auditoria
// e a atualizacao anual das tabelas legais.
//
// Vigencia das tabelas:
//   - INSS: valores em vigencia a partir de 01/01/2024 (Portaria MPS/ME n. 2.017/2023):
//     limites 1412.00, 2666.68, 4000.03, 7786.02 com aliquota progressiva 7,5%/9%/12%/14%.
//   - IRRF: tabela em vigencia a partir de 01/05/2023 (Lei n. 14.713/2023), aplicada a partir de 2024:
//     isencao ate 2259.20 e limites 2826.65, 3751.05, 4664.68 com deducoes 169.44, 353.44, 634.77, 869.36.
//
// O comportamento numerico de todas as funcoes e identico ao do periodo em que os literais eram
// hardcoded. Alterar estes valores requer revisao dos testes de dominio (golden).

// taxBracket representa uma faixa de contribuicao progressiva do INSS.
type taxBracket struct {
	limit float64
	rate  float64
}

// inssBrackets define as faixas progressivas de contribuicao do INSS.
var inssBrackets = []taxBracket{
	{limit: 1412.00, rate: 0.075},
	{limit: 2666.68, rate: 0.09},
	{limit: 4000.03, rate: 0.12},
	{limit: 7786.02, rate: 0.14},
}

// irrfBracket representa uma faixa progressiva do IRRF.
type irrfBracket struct {
	limit     float64
	rate      float64
	deduction float64
}

// IRRFExemptionLimit e o limite de isencao do IRRF sobre a base de calculo (salario bruto - INSS).
const IRRFExemptionLimit = 2259.20

// irrfBrackets define as faixas progressivas do IRRF (base = salario bruto - INSS).
var irrfBrackets = []irrfBracket{
	{limit: 2826.65, rate: 0.075, deduction: 169.44},
	{limit: 3751.05, rate: 0.15, deduction: 353.44},
	{limit: 4664.68, rate: 0.225, deduction: 634.77},
	{limit: math.Inf(1), rate: 0.275, deduction: 869.36},
}

// Limiares das zonas de decisao do IDT (percentual).
const (
	ZoneIDTGreenThreshold = 10.0
	ZoneIDTRedThreshold   = 15.0
)

// Limiares das zonas de diagnostico do TCM (percentual).
const (
	ZoneTCMHighThreshold    = 45.0
	ZoneTCMAlertThreshold   = 55.0
	ZoneTCMAnomalyThreshold = 60.0
)

// Dias trabalhados medios mensais estimados por escala de trabalho (usados quando nao ha calculo exato).
const (
	ScheduleShiftDays5x2   = 22.0
	ScheduleShiftDays6x1   = 26.0
	ScheduleShiftDays12x36 = 15.0
	ScheduleShiftDays4x3   = 17.0
)
