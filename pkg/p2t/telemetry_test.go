package p2t_test

import (
	"math"
	"testing"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
)

const floatTolerance = 0.0001

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < floatTolerance
}

func TestCalculateTelemetry_HappyPath(t *testing.T) {
	input := p2t.TelemetryInput{
		GrossSalary:     5000.0,
		FixedDeductions: 800.0,
		ErrorDeductions: 200.0,
		InvisibleCosts:  300.0,
		ContractHours:   160.0,
		CommuteHours:    40.0,
	}

	result, err := p2t.CalculateTelemetry(input)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// HT = 160 + 40 = 200
	if !almostEqual(result.TotalHours, 200.0) {
		t.Errorf("esperado HT=200.0, obtido: %.2f", result.TotalHours)
	}

	// SL = 5000 - (800 + 200 + 300) = 3700
	if !almostEqual(result.RealLiquidity, 3700.0) {
		t.Errorf("esperado SL=3700.0, obtido: %.2f", result.RealLiquidity)
	}

	// VRH = 3700 / 200 = 18.5
	if !almostEqual(result.VRH, 18.5) {
		t.Errorf("esperado VRH=18.5, obtido: %.2f", result.VRH)
	}

	// IDT = ((200 + 300) / 5000) * 100 = 10.0%
	if !almostEqual(result.IDT, 10.0) {
		t.Errorf("esperado IDT=10.0, obtido: %.2f", result.IDT)
	}
}

func TestCalculateTelemetry_ValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input p2t.TelemetryInput
	}{
		{
			name: "GrossSalary Zero",
			input: p2t.TelemetryInput{
				GrossSalary:   0,
				ContractHours: 160,
			},
		},
		{
			name: "ContractHours Zero",
			input: p2t.TelemetryInput{
				GrossSalary:   5000,
				ContractHours: 0,
			},
		},
		{
			name: "Negative Fixed Deductions",
			input: p2t.TelemetryInput{
				GrossSalary:     5000,
				ContractHours:   160,
				FixedDeductions: -100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p2t.CalculateTelemetry(tt.input)
			if err == nil {
				t.Errorf("esperava erro de validacao para %s, mas obteve nil", tt.name)
			}
		})
	}
}

func TestCalculateIDT3(t *testing.T) {
	t.Run("Valido com 3+ elementos", func(t *testing.T) {
		history := []float64{12.0, 8.0, 9.0, 11.0, 10.0} // ultimos 3: 9.0, 11.0, 10.0
		idt3, err := p2t.CalculateIDT3(history)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		expected := (9.0 + 11.0 + 10.0) / 3.0 // 10.0
		if !almostEqual(idt3, expected) {
			t.Errorf("esperado IDT3=%.2f, obtido=%.2f", expected, idt3)
		}
	})

	t.Run("Historico insuficiente", func(t *testing.T) {
		history := []float64{10.0, 12.0}
		_, err := p2t.CalculateIDT3(history)
		if err == nil {
			t.Error("esperava erro para historico menor que 3 elementos")
		}
	})
}

func TestEvaluateIDTZone(t *testing.T) {
	tests := []struct {
		idt3         float64
		expectedZone p2t.IDTZone
	}{
		{5.0, p2t.ZoneGreen},
		{9.99, p2t.ZoneGreen},
		{10.0, p2t.ZoneYellow},
		{14.99, p2t.ZoneYellow},
		{15.0, p2t.ZoneRed},
		{25.0, p2t.ZoneRed},
	}

	for _, tt := range tests {
		zone, desc := p2t.EvaluateIDTZone(tt.idt3)
		if zone != tt.expectedZone {
			t.Errorf("para IDT3=%.2f esperado zone %s, obtido %s", tt.idt3, tt.expectedZone, zone)
		}
		if desc == "" {
			t.Errorf("descricao da zona nao deve ser vazia")
		}
	}
}

func TestCalculateCommuteHours(t *testing.T) {
	tests := []struct {
		name          string
		schedule      p2t.WorkSchedule
		dailyHours    float64
		expectedHours float64
		expectErr     bool
	}{
		{"5x2 com 1.5h/dia", p2t.Schedule5x2, 1.5, 33.0, false},
		{"6x1 com 1.5h/dia", p2t.Schedule6x1, 1.5, 39.0, false},
		{"12x36 com 1.5h/dia", p2t.Schedule12x36, 1.5, 22.5, false},
		{"4x3 com 2.0h/dia", p2t.Schedule4x3, 2.0, 34.0, false},
		{"Horas Negativas", p2t.Schedule5x2, -1.0, 0, true},
		{"Escala Invalida", p2t.WorkSchedule("99x99"), 1.0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := p2t.CalculateCommuteHours(tt.schedule, tt.dailyHours)
			if tt.expectErr {
				if err == nil {
					t.Errorf("esperado erro para %s, obtido nil", tt.name)
				}
			} else {
				if err != nil {
					t.Fatalf("erro inesperado para %s: %v", tt.name, err)
				}
				if !almostEqual(h, tt.expectedHours) {
					t.Errorf("esperado %.2f h, obtido %.2f h", tt.expectedHours, h)
				}
			}
		})
	}
}

func TestCalculateExactShifts(t *testing.T) {
	// Em Julho/2026 (31 dias): 1/Jul eh quarta-feira.
	// Se refDate = 01/Jul/2026 (plantao no dia 1): plantoes nos dias 1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31 = 16 plantoes.
	refDate1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	shifts16, err := p2t.CalculateExactShifts(p2t.Schedule12x36, refDate1, "2026-07")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if shifts16 != 16 {
		t.Errorf("esperado 16 plantoes em julho começando no dia 1, obtido: %d", shifts16)
	}

	// Se refDate = 02/Jul/2026 (folga no dia 1, plantao no dia 2): plantoes nos dias 2, 4, 6... 30 = 15 plantoes.
	refDate2 := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	shifts15, err := p2t.CalculateExactShifts(p2t.Schedule12x36, refDate2, "2026-07")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if shifts15 != 15 {
		t.Errorf("esperado 15 plantoes em julho começando no dia 2, obtido: %d", shifts15)
	}

	// Testar CalculateCommuteHoursWithRefDate
	commuteHours, count, err := p2t.CalculateCommuteHoursWithRefDate(p2t.Schedule12x36, 2.0, refDate1, "2026-07")
	if err != nil {
		t.Fatalf("erro no calculo de commute com refDate: %v", err)
	}
	if count != 16 || !almostEqual(commuteHours, 32.0) {
		t.Errorf("esperado 16 plantoes e 32.0h, obtido %d plantoes e %.2fh", count, commuteHours)
	}
}


