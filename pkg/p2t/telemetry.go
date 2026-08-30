package p2t

import (
	"errors"
	"fmt"
	"time"
)

// IDTZone representa a zona de decisao baseada no IDT.
type IDTZone string

const (
	ZoneGreen  IDTZone = "Zona Verde"
	ZoneYellow IDTZone = "Zona Amarela"
	ZoneRed    IDTZone = "Zona Vermelha"
)

// WorkSchedule representa o tipo de escala de trabalho presencial.
type WorkSchedule string

const (
	Schedule5x2   WorkSchedule = "5x2"
	Schedule6x1   WorkSchedule = "6x1"
	Schedule12x36 WorkSchedule = "12x36"
	Schedule4x3   WorkSchedule = "4x3"
)

// CalculateExactShifts calcula a quantidade exata de plantões/dias trabalhados no mês de competência a partir de uma data de referência.
func CalculateExactShifts(schedule WorkSchedule, refDate time.Time, referenceMonth string) (int, error) {
	parsedMonth, err := time.Parse("2006-01", referenceMonth)
	if err != nil {
		return 0, fmt.Errorf("formato de mês de competência inválido '%s': esperado YYYY-MM", referenceMonth)
	}

	startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	refDateUTC := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)

	shifts := 0
	for d := startOfMonth; !d.After(endOfMonth); d = d.AddDate(0, 0, 1) {
		switch schedule {
		case Schedule12x36:
			diffDays := int(d.Sub(refDateUTC).Hours() / 24)
			if diffDays%2 == 0 {
				shifts++
			}
		case Schedule5x2:
			if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
				shifts++
			}
		case Schedule6x1:
			if d.Weekday() != time.Sunday {
				shifts++
			}
		case Schedule4x3:
			if d.Weekday() >= time.Monday && d.Weekday() <= time.Thursday {
				shifts++
			}
		default:
			return 0, fmt.Errorf("escala de trabalho inválida ou não suportada: '%s'", schedule)
		}
	}

	return shifts, nil
}

// CalculateCommuteHoursWithRefDate calcula as horas de deslocamento considerando o calendário exato do mês.
func CalculateCommuteHoursWithRefDate(schedule WorkSchedule, dailyCommuteHours float64, refDate time.Time, referenceMonth string) (float64, int, error) {
	if dailyCommuteHours < 0 {
		return 0, 0, errors.New("horas de deslocamento diario nao podem ser negativas")
	}

	exactShifts, err := CalculateExactShifts(schedule, refDate, referenceMonth)
	if err != nil {
		return 0, 0, err
	}

	totalHours := RoundCurrency(float64(exactShifts) * dailyCommuteHours)
	return totalHours, exactShifts, nil
}

// CalculateCommuteHours calcula o total de horas mensais de deslocamento (HD) com base na escala estimada e horas diarias de trânsito.
func CalculateCommuteHours(schedule WorkSchedule, dailyCommuteHours float64) (float64, error) {
	if dailyCommuteHours < 0 {
		return 0, errors.New("horas de deslocamento diario nao podem ser negativas")
	}

	var monthlyDays float64
	switch schedule {
	case Schedule5x2:
		monthlyDays = ScheduleShiftDays5x2
	case Schedule6x1:
		monthlyDays = ScheduleShiftDays6x1
	case Schedule12x36:
		monthlyDays = ScheduleShiftDays12x36
	case Schedule4x3:
		monthlyDays = ScheduleShiftDays4x3
	default:
		return 0, fmt.Errorf("escala de trabalho invalida ou nao suportada: '%s'", schedule)
	}

	return RoundCurrency(monthlyDays * dailyCommuteHours), nil
}

// TelemetryInput engloba as variaveis de um ciclo mensal para telemetria de tempo e retorno.
type TelemetryInput struct {
	GrossSalary     float64 // SB: Salario Bruto
	FixedDeductions float64 // DF: Descontos Legais Fixos
	ErrorDeductions float64 // DE: Descontos decorrentes de erros operacionais
	InvisibleCosts  float64 // CI: Custos invisiveis do bolso
	ContractHours   float64 // HC: Horas contratualmente devidas
	CommuteHours    float64 // HD: Horas mensais de deslocamento
}

// TelemetryResult contem as metricas deduzidas para o ciclo.
type TelemetryResult struct {
	TotalHours    float64 // HT = HC + HD
	RealLiquidity float64 // SL = SB - (DF + DE + CI)
	VRH           float64 // VRH = SL / HT
	IDT           float64 // IDT = ((DE + CI) / SB) * 100
}

// Validate verifica a consistencia dos parametros de entrada.
func (t TelemetryInput) Validate() error {
	if t.GrossSalary <= 0 {
		return errors.New("salario bruto (SB) deve ser maior que zero")
	}
	if t.ContractHours <= 0 {
		return errors.New("horas contratuais (HC) devem ser maiores que zero")
	}
	if t.FixedDeductions < 0 || t.ErrorDeductions < 0 || t.InvisibleCosts < 0 || t.CommuteHours < 0 {
		return errors.New("descontos e horas de deslocamento nao podem ser negativos")
	}
	return nil
}

// CalculateTelemetry calcula HT, SL, VRH e IDT a partir de um TelemetryInput com precisao monetaria.
func CalculateTelemetry(input TelemetryInput) (TelemetryResult, error) {
	if err := input.Validate(); err != nil {
		return TelemetryResult{}, err
	}

	ht := input.ContractHours + input.CommuteHours
	sl := RoundCurrency(input.GrossSalary - (input.FixedDeductions + input.ErrorDeductions + input.InvisibleCosts))
	vrh := RoundCurrency(sl / ht)
	idt := RoundCurrency(((input.ErrorDeductions + input.InvisibleCosts) / input.GrossSalary) * 100.0)

	return TelemetryResult{
		TotalHours:    ht,
		RealLiquidity: sl,
		VRH:           vrh,
		IDT:           idt,
	}, nil
}

// CalculateIDT3 calcula a media movel de 3 meses dos valores de IDT com precisao.
func CalculateIDT3(idtHistory []float64) (float64, error) {
	if len(idtHistory) < 3 {
		return 0, fmt.Errorf("historico insuficiente para IDT3: esperado no minimo 3 registros, recebido %d", len(idtHistory))
	}
	// Considera os 3 ultimos registros
	recent := idtHistory[len(idtHistory)-3:]
	sum := recent[0] + recent[1] + recent[2]
	return RoundCurrency(sum / 3.0), nil
}

// EvaluateIDTZone determina a zona de decisao com base no IDT3 ou IDT atual.
func EvaluateIDTZone(idt3 float64) (IDTZone, string) {
	idt3 = RoundCurrency(idt3)
	switch {
	case idt3 < ZoneIDTGreenThreshold:
		return ZoneGreen, "Estavel / Padrao Operacional"
	case idt3 >= ZoneIDTGreenThreshold && idt3 < ZoneIDTRedThreshold:
		return ZoneYellow, "Alerta de Corrosao / Ativar busca passiva de vagas"
	default:
		return ZoneRed, "Inviabilidade Financeira / Gatilho de Saida Ativado"
	}
}
