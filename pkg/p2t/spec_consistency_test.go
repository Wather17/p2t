package p2t_test

import (
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
)

// TestSpecExampleTelemetry reproduz o exemplo da seção 1.4 de docs/specification.md
// (SB 5000, DF 800, DE 0, CI 200, HC 160, HD 33 => HT=193, SL=4000, VRH=20.73, IDT=4.00) e
// falha se a implementacao divergir da especificacao.
func TestSpecExampleTelemetry(t *testing.T) {
	input := p2t.TelemetryInput{
		GrossSalary:     5000.00, // SB
		FixedDeductions: 800.00,  // DF
		ErrorDeductions: 0.00,    // DE
		InvisibleCosts:  200.00,  // CI
		ContractHours:   160.00,  // HC
		CommuteHours:    33.00,   // HD
	}

	res, err := p2t.CalculateTelemetry(input)
	if err != nil {
		t.Fatalf("erro inesperado no calculo: %v", err)
	}

	if res.TotalHours != 193.00 { // HT = HC + HD
		t.Errorf("HT esperado 193.00, obtido %.2f", res.TotalHours)
	}
	if res.RealLiquidity != 4000.00 { // SL = SB - (DF + DE + CI)
		t.Errorf("SL esperado 4000.00, obtido %.2f", res.RealLiquidity)
	}
	if !almostEqual(res.VRH, 20.73) { // VRH = SL / HT
		t.Errorf("VRH esperado 20.73, obtido %.4f", res.VRH)
	}
	if !almostEqual(res.IDT, 4.00) { // IDT = ((DE + CI) / SB) * 100
		t.Errorf("IDT esperado 4.00, obtido %.4f", res.IDT)
	}

	// IDT3: com os tres ultimos IDTs 4.00 a media e 4.00 (secao 1.5)
	idt3, err := p2t.CalculateIDT3([]float64{res.IDT, res.IDT, res.IDT})
	if err != nil {
		t.Fatalf("erro no calculo de IDT3: %v", err)
	}
	if !almostEqual(idt3, 4.00) {
		t.Errorf("IDT3 esperado 4.00, obtido %.4f", idt3)
	}
}

// TestIDTZoneBoundaries valida os limiares das zonas de decisao do IDT (spec secao 1.5):
// < 10 verde, 10-15 amarela, >= 15 vermelha.
func TestIDTZoneBoundaries(t *testing.T) {
	cases := []struct {
		idt  float64
		zone p2t.IDTZone
	}{
		{9.99, p2t.ZoneGreen},
		{10.0, p2t.ZoneYellow},
		{14.99, p2t.ZoneYellow},
		{15.0, p2t.ZoneRed},
	}

	for _, c := range cases {
		zone, _ := p2t.EvaluateIDTZone(c.idt)
		if zone != c.zone {
			t.Errorf("para IDT %.2f esperado %s, obtido %s", c.idt, c.zone, zone)
		}
	}
}

// TestEfficiencyZoneBoundaries valida os limiares das faixas de eficiencia do TCM (spec secao 2.3):
// <45 alta, 45-55 estavel, >55-60 alerta, >60 anomalia.
func TestEfficiencyZoneBoundaries(t *testing.T) {
	cases := []struct {
		tcm    float64
		expect p2t.EfficiencyDiagnosis
	}{
		{44.99, p2t.HighEfficiency},
		{45.00, p2t.StableEfficiency},
		{55.00, p2t.StableEfficiency},
		{55.01, p2t.AlertEfficiency},
		{60.00, p2t.AlertEfficiency},
		{60.01, p2t.ConsumptionAnomaly},
	}

	for _, c := range cases {
		diagnosis, _ := p2t.DiagnoseEfficiency(c.tcm)
		if diagnosis != c.expect {
			t.Errorf("para TCM %.2f esperado %s, obtido %s", c.tcm, c.expect, diagnosis)
		}
	}
}
