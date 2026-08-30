package storage

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wather17/p2t/pkg/p2t"
)

// exportChunkSize define o tamanho de cada lote de leitura paginada no export.
const exportChunkSize = 1000

// GetAllTelemetryRecords percorre o historico completo com paginacao por id (chunkSize) e retorna tudo em ordem ascendente.
func GetAllTelemetryRecords(repo *Repository) ([]TelemetryRecord, error) {
	var records []TelemetryRecord
	afterID := int64(0)
	for {
		chunk, err := repo.GetTelemetryRecordsChunk(afterID, exportChunkSize)
		if err != nil {
			return nil, fmt.Errorf("falha ao obter registros para exportar: %w", err)
		}
		records = append(records, chunk...)
		if len(chunk) < exportChunkSize {
			break
		}
		afterID = chunk[len(chunk)-1].ID
	}
	return records, nil
}

// ExportTelemetryJSON exporta todos os registros de telemetria para o formato JSON.
func ExportTelemetryJSON(repo *Repository) ([]byte, error) {
	records, err := GetAllTelemetryRecords(repo)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar para JSON: %w", err)
	}

	return data, nil
}

// ExportTelemetryCSV exporta todos os registros de telemetria para o formato CSV.
func ExportTelemetryCSV(repo *Repository) ([]byte, error) {
	records, err := GetAllTelemetryRecords(repo)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Cabeçalho CSV
	header := []string{
		"id", "reference_month", "gross_salary", "fixed_deductions", "error_deductions",
		"invisible_costs", "contract_hours", "commute_hours", "total_hours",
		"real_liquidity", "vrh", "idt", "created_at",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, r := range records {
		row := []string{
			strconv.FormatInt(r.ID, 10),
			r.ReferenceMonth,
			fmt.Sprintf("%.2f", r.GrossSalary),
			fmt.Sprintf("%.2f", r.FixedDeductions),
			fmt.Sprintf("%.2f", r.ErrorDeductions),
			fmt.Sprintf("%.2f", r.InvisibleCosts),
			fmt.Sprintf("%.2f", r.ContractHours),
			fmt.Sprintf("%.2f", r.CommuteHours),
			fmt.Sprintf("%.2f", r.TotalHours),
			fmt.Sprintf("%.2f", r.RealLiquidity),
			strconv.FormatFloat(r.VRH, 'f', 2, 64),
			strconv.FormatFloat(r.IDT, 'f', 2, 64),
			r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ImportResult descreve o resultado de uma importacao: o que entrou, foi pulado por duplicidade ou falhou,
// com motivos resumidos para o usuario.
type ImportResult struct {
	Inserted   int
	Duplicated int
	Failed     int
	Why        []string
}

func (r *ImportResult) why(reason string) {
	if len(r.Why) < 10 {
		r.Why = append(r.Why, reason)
	}
}

// ImportTelemetryJSON importa registros de telemetria a partir de um JSON.
// Nenhuma linha e contabilizada sem motivo: as causas de falha entram em Why (limitado a 10).
func ImportTelemetryJSON(repo *Repository, data []byte) (ImportResult, error) {
	var records []TelemetryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return ImportResult{}, fmt.Errorf("falha ao decodificar JSON de telemetria: %w", err)
	}

	result := ImportResult{}
	for i, r := range records {
		input := p2t.TelemetryInput{
			GrossSalary:     r.GrossSalary,
			FixedDeductions: r.FixedDeductions,
			ErrorDeductions: r.ErrorDeductions,
			InvisibleCosts:  r.InvisibleCosts,
			ContractHours:   r.ContractHours,
			CommuteHours:    r.CommuteHours,
		}
		res, err := p2t.CalculateTelemetry(input)
		if err != nil {
			result.Failed++
			result.why(fmt.Sprintf("registro [%d]: %v", i, err))
			continue
		}

		ref := r.ReferenceMonth
		if ref == "" {
			ref = r.CreatedAt.Format("2006-01")
		}

		if _, err := repo.SaveTelemetry(input, res, ref); err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				result.Duplicated++
				result.why(fmt.Sprintf("registro [%d]: competência %s já existe no banco", i, ref))
			} else {
				result.Failed++
				result.why(fmt.Sprintf("registro [%d]: %v", i, err))
			}
			continue
		}
		result.Inserted++
	}

	return result, nil
}

// ImportTelemetryCSV importa registros de telemetria a partir de um CSV.
// Linhas curtas ou com numeros malformados contam como Failed com o numero da linha e nada e inserido delas.
func ImportTelemetryCSV(repo *Repository, data []byte) (ImportResult, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return ImportResult{}, fmt.Errorf("falha ao ler CSV de telemetria: %w", err)
	}

	if len(rows) < 2 {
		return ImportResult{}, nil
	}

	result := ImportResult{}
	// Pula cabeçalho (linha 0)
	for i, row := range rows[1:] {
		line := i + 2
		if len(row) < 13 {
			result.Failed++
			result.why(fmt.Sprintf("linha %d: esperadas 13 colunas, obtidas %d", line, len(row)))
			continue
		}

		refMonth := row[1]
		vals := []float64{}
		fields := []string{row[2], row[3], row[4], row[5], row[6], row[7]}
		parseErr := false
		for _, f := range fields {
			v, err := strconv.ParseFloat(strings.TrimSpace(f), 64)
			if err != nil {
				parseErr = true
				break
			}
			vals = append(vals, v)
		}
		if parseErr {
			result.Failed++
			result.why(fmt.Sprintf("linha %d: campo numerico malformado", line))
			continue
		}

		input := p2t.TelemetryInput{
			GrossSalary:     vals[0],
			FixedDeductions: vals[1],
			ErrorDeductions: vals[2],
			InvisibleCosts:  vals[3],
			ContractHours:   vals[4],
			CommuteHours:    vals[5],
		}

		res, err := p2t.CalculateTelemetry(input)
		if err != nil {
			result.Failed++
			result.why(fmt.Sprintf("linha %d: %v", line, err))
			continue
		}

		if _, err := repo.SaveTelemetry(input, res, refMonth); err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				result.Duplicated++
				result.why(fmt.Sprintf("linha %d: competência %s já existe no banco", line, refMonth))
			} else {
				result.Failed++
				result.why(fmt.Sprintf("linha %d: %v", line, err))
			}
			continue
		}
		result.Inserted++
	}

	return result, nil
}
