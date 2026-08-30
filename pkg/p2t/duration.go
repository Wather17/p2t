package p2t

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// "1h30", "1 h30", "1h 30min", "1,5h", "1.5h"
	reHours = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*h\s*(?:([0-9]+)\s*(?:min)?)?$`)
	// "90min", "75m", "45 minutos", "90 min"
	reMinutes = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*(?:min|m|minuto|minutos)$`)
	// numero puro/decimial
	reNumber = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?$`)
)

// ParseHumanDurationMinutes converte duracoes humanas de tempo em minutos (float64 com 2 casas, via RoundCurrency).
//
// Regras de desambiguacao:
//   - numero puro e interpretado como minutos ("90" -> 90.0, "0" -> 0.0);
//   - "h" (horas) com minutos opcionais: "1h" -> 60.0, "1h30" -> 90.0, "1,5h" -> 90.0, "1h30min" -> 90.0;
//   - "min"/"m"/"minuto(s)": "90min" -> 90.0, "45 minutos" -> 45.0;
//   - "H:MM" (com "h" opcional no final): "1:30" -> 90.0, "1:30h" -> 90.0;
//   - espacos entre valor e unidade aceitos ("1 h30", "90 min");
//   - valor negativo, moeda ("R$ 90"), segundos/misturas ("1h30min45s") e texto invalido retornam erro.
func ParseHumanDurationMinutes(input string) (float64, error) {
	str := strings.TrimSpace(input)
	if str == "" {
		return 0, fmt.Errorf("duracao vazia: informe um tempo como '1h30', '90min' ou '45'")
	}

	lower := strings.ToLower(input)
	if strings.Contains(lower, "r$") || strings.Contains(lower, "$") {
		return 0, fmt.Errorf("duracao '%s' nao pode conter valores monetarios", input)
	}

	str = strings.ReplaceAll(str, ",", ".")
	str = strings.TrimSpace(str)

	// H:MM (horas:minutos), com "h" opcional: "1:30" / "1:30h"
	if strings.Contains(str, ":") {
		work := strings.TrimSpace(strings.TrimSuffix(str, "h"))
		parts := strings.Split(work, ":")
		if len(parts) != 2 {
			return 0, fmt.Errorf("duracao '%s' invalida: use H:MM (ex.: 1:30 = 1h30min)", input)
		}
		h, errH := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		m, errM := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if errH != nil || errM != nil || h < 0 || m < 0 || m >= 60 {
			return 0, fmt.Errorf("duracao '%s' invalida: horas e minutos devem ser validos (minutos < 60)", input)
		}
		return RoundCurrency(h*60.0 + m), nil
	}

	// "1h", "1h30", "1,5h", "1h 30min"
	if m := reHours.FindStringSubmatch(str); m != nil {
		h, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, fmt.Errorf("duracao '%s' invalida: horas nao numericas", input)
		}
		min := 0.0
		if m[2] != "" {
			if min, err = strconv.ParseFloat(m[2], 64); err != nil {
				return 0, fmt.Errorf("duracao '%s' invalida: minutos nao numericos", input)
			}
		}
		return RoundCurrency(h*60.0 + min), nil
	}

	// "90min", "75m", "45 minutos"
	if m := reMinutes.FindStringSubmatch(str); m != nil {
		min, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, fmt.Errorf("duracao '%s' invalida: minutos nao numericos", input)
		}
		return RoundCurrency(min), nil
	}

	// numero puro = minutos
	if reNumber.MatchString(str) {
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, fmt.Errorf("duracao '%s' invalida: valor nao numerico", input)
		}
		return RoundCurrency(num), nil
	}

	return 0, fmt.Errorf("duracao '%s' invalida: informe um tempo como '1h30', '90min', '1:30' ou '90'", input)
}
