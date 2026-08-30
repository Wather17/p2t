package storage

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"

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

// ImportTelemetryJSON importa registros de telemetria a partir de um JSON.
func ImportTelemetryJSON(repo *Repository, data []byte) (int, error) {
	var records []TelemetryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return 0, fmt.Errorf("falha ao decodificar JSON de telemetria: %w", err)
	}

	imported := 0
	for _, r := range records {
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
			continue
		}

		ref := r.ReferenceMonth
		if ref == "" {
			ref = r.CreatedAt.Format("2006-01")
		}

		if _, err := repo.SaveTelemetry(input, res, ref); err == nil {
			imported++
		}
	}

	return imported, nil
}

// ImportTelemetryCSV importa registros de telemetria a partir de um CSV.
func ImportTelemetryCSV(repo *Repository, data []byte) (int, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	rows, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("falha ao ler CSV de telemetria: %w", err)
	}

	if len(rows) < 2 {
		return 0, nil
	}

	imported := 0
	// Pula cabeçalho (linha 0)
	for _, row := range rows[1:] {
		if len(row) < 12 {
			continue
		}

		refMonth := row[1]
		grossSalary, _ := strconv.ParseFloat(row[2], 64)
		fixedDeductions, _ := strconv.ParseFloat(row[3], 64)
		errorDeductions, _ := strconv.ParseFloat(row[4], 64)
		invisibleCosts, _ := strconv.ParseFloat(row[5], 64)
		contractHours, _ := strconv.ParseFloat(row[6], 64)
		commuteHours, _ := strconv.ParseFloat(row[7], 64)

		input := p2t.TelemetryInput{
			GrossSalary:     grossSalary,
			FixedDeductions: fixedDeductions,
			ErrorDeductions: errorDeductions,
			InvisibleCosts:  invisibleCosts,
			ContractHours:   contractHours,
			CommuteHours:    commuteHours,
		}

		res, err := p2t.CalculateTelemetry(input)
		if err != nil {
			continue
		}

		if _, err := repo.SaveTelemetry(input, res, refMonth); err == nil {
			imported++
		}
	}

	return imported, nil
}
