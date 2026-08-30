package p2t_test

import (
	"testing"

	"github.com/Wather17/p2t/pkg/p2t"
)

func TestParseHumanDurationMinutes(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"90", 90.0},
		{"90min", 90.0},
		{"75m", 75.0},
		{"45 minutos", 45.0},
		{"90 min", 90.0},
		{"1h", 60.0},
		{"1h30", 90.0},
		{"1h30min", 90.0},
		{"1 h30", 90.0},
		{"1,5h", 90.0},
		{"1.5h", 90.0},
		{"2h", 120.0},
		{"1:30", 90.0},
		{"1:30h", 90.0},
		{"0", 0.0},
		{"1h20", 80.0},
	}

	for _, tt := range tests {
		got, err := p2t.ParseHumanDurationMinutes(tt.input)
		if err != nil {
			t.Errorf("erro inesperado para input %q: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("para input %q esperado %.2f, obtido %.2f", tt.input, tt.expected, got)
		}
	}
}

func TestParseHumanDurationMinutes_Errors(t *testing.T) {
	inputs := []string{"", "abc", "-30min", "R$ 90", "1h30min45s", "1:60", "1:30:00"}

	for _, input := range inputs {
		if _, err := p2t.ParseHumanDurationMinutes(input); err == nil {
			t.Errorf("esperado erro para input %q, obtido nil", input)
		}
	}
}
