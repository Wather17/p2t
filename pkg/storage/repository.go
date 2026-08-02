package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
)

// TelemetryRecord representa um ciclo de telemetria armazenado.
type TelemetryRecord struct {
	ID              int64
	GrossSalary     float64
	FixedDeductions float64
	ErrorDeductions float64
	InvisibleCosts  float64
	ContractHours   float64
	CommuteHours    float64
	TotalHours      float64
	RealLiquidity   float64
	VRH             float64
	IDT             float64
	ReferenceMonth  string
	CreatedAt       time.Time
}

// BufferRecord representa um registro de ciclo da Caixinha armazenado.
type BufferRecord struct {
	ID               int64
	Cap              float64
	RemainingBalance float64
	InvisibleCost    float64
	CreatedAt        time.Time
}

// Repository lida com as operacoes de persistencia no SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository cria um novo repositorio vinculado a conexao SQLite.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveTelemetry insere ou atualiza um registro de telemetria no banco.
func (r *Repository) SaveTelemetry(input p2t.TelemetryInput, res p2t.TelemetryResult, referenceMonth string) (int64, error) {
	query := `
	INSERT INTO telemetry_cycles (
		gross_salary, fixed_deductions, error_deductions, invisible_costs,
		contract_hours, commute_hours, total_hours, real_liquidity, vrh, idt, reference_month
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	result, err := r.db.Exec(query,
		input.GrossSalary, input.FixedDeductions, input.ErrorDeductions, input.InvisibleCosts,
		input.ContractHours, input.CommuteHours, res.TotalHours, res.RealLiquidity, res.VRH, res.IDT, referenceMonth,
	)
	if err != nil {
		return 0, fmt.Errorf("falha ao salvar telemetria: %w", err)
	}

	return result.LastInsertId()
}

// GetRecentIDTHistory retorna os N IDTs mais recentes ordenados cronologicamente.
func (r *Repository) GetRecentIDTHistory(limit int) ([]float64, error) {
	query := `
	SELECT idt FROM (
		SELECT idt, id FROM telemetry_cycles ORDER BY id DESC LIMIT ?
	) ORDER BY id ASC;
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar historico de IDT: %w", err)
	}
	defer rows.Close()

	var history []float64
	for rows.Next() {
		var idt float64
		if err := rows.Scan(&idt); err != nil {
			return nil, err
		}
		history = append(history, idt)
	}
	return history, rows.Err()
}

// GetRecentTelemetryRecords retorna os N registros de telemetria mais recentes ordenados cronologicamente.
func (r *Repository) GetRecentTelemetryRecords(limit int) ([]TelemetryRecord, error) {
	query := `
	SELECT id, gross_salary, fixed_deductions, error_deductions, invisible_costs,
	       contract_hours, commute_hours, total_hours, real_liquidity, vrh, idt, reference_month, created_at
	FROM (
		SELECT id, gross_salary, fixed_deductions, error_deductions, invisible_costs,
		       contract_hours, commute_hours, total_hours, real_liquidity, vrh, idt, reference_month, created_at
		FROM telemetry_cycles ORDER BY id DESC LIMIT ?
	) ORDER BY id ASC;
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar registros de telemetria: %w", err)
	}
	defer rows.Close()

	var records []TelemetryRecord
	for rows.Next() {
		var rec TelemetryRecord
		err := rows.Scan(
			&rec.ID, &rec.GrossSalary, &rec.FixedDeductions, &rec.ErrorDeductions, &rec.InvisibleCosts,
			&rec.ContractHours, &rec.CommuteHours, &rec.TotalHours, &rec.RealLiquidity, &rec.VRH, &rec.IDT, &rec.ReferenceMonth, &rec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// GetLatestTelemetryRecord retorna o registro de telemetria mais recente salvo no banco, se existir.
func (r *Repository) GetLatestTelemetryRecord() (*TelemetryRecord, error) {
	query := `
	SELECT id, gross_salary, fixed_deductions, error_deductions, invisible_costs,
	       contract_hours, commute_hours, total_hours, real_liquidity, vrh, idt, reference_month, created_at
	FROM telemetry_cycles ORDER BY id DESC LIMIT 1;
	`
	var rec TelemetryRecord
	err := r.db.QueryRow(query).Scan(
		&rec.ID, &rec.GrossSalary, &rec.FixedDeductions, &rec.ErrorDeductions, &rec.InvisibleCosts,
		&rec.ContractHours, &rec.CommuteHours, &rec.TotalHours, &rec.RealLiquidity, &rec.VRH, &rec.IDT, &rec.ReferenceMonth, &rec.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar ultimo registro de telemetria: %w", err)
	}
	return &rec, nil
}

// GetTelemetryByReferenceMonth busca um registro de telemetria pelo mes de competencia (YYYY-MM).
func (r *Repository) GetTelemetryByReferenceMonth(refMonth string) (*TelemetryRecord, error) {
	query := `
	SELECT id, gross_salary, fixed_deductions, error_deductions, invisible_costs,
	       contract_hours, commute_hours, total_hours, real_liquidity, vrh, idt, reference_month, created_at
	FROM telemetry_cycles WHERE reference_month = ?;
	`
	var rec TelemetryRecord
	err := r.db.QueryRow(query, refMonth).Scan(
		&rec.ID, &rec.GrossSalary, &rec.FixedDeductions, &rec.ErrorDeductions, &rec.InvisibleCosts,
		&rec.ContractHours, &rec.CommuteHours, &rec.TotalHours, &rec.RealLiquidity, &rec.VRH, &rec.IDT, &rec.ReferenceMonth, &rec.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar telemetria por competencia %s: %w", refMonth, err)
	}
	return &rec, nil
}



// SaveBufferCycle insere um registro de buffer operacional.
func (r *Repository) SaveBufferCycle(cap, remainingBalance, invisibleCost float64) (int64, error) {
	query := `
	INSERT INTO buffer_cycles (cap, remaining_balance, invisible_cost)
	VALUES (?, ?, ?);
	`
	result, err := r.db.Exec(query, cap, remainingBalance, invisibleCost)
	if err != nil {
		return 0, fmt.Errorf("falha ao salvar ciclo de buffer: %w", err)
	}
	return result.LastInsertId()
}

// GetRecentReplenishments retorna os custos invisiveis (reposicoes) dos ultimos N ciclos do buffer.
func (r *Repository) GetRecentReplenishments(limit int) ([]float64, error) {
	query := `
	SELECT invisible_cost FROM (
		SELECT invisible_cost, id FROM buffer_cycles ORDER BY id DESC LIMIT ?
	) ORDER BY id ASC;
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar reposicoes do buffer: %w", err)
	}
	defer rows.Close()

	var list []float64
	for rows.Next() {
		var cost float64
		if err := rows.Scan(&cost); err != nil {
			return nil, err
		}
		list = append(list, cost)
	}
	return list, rows.Err()
}

// TelemetryAnalytics agrega métricas estatísticas do histórico.
type TelemetryAnalytics struct {
	IDT3                      float64
	IDT6                      float64
	IDT12                     float64
	VRHTrendPercent           float64
	TotalAnnualInvisibleCosts float64
	TotalRecords              int
}

// GetTelemetryAnalytics calcula estatisticas agregadas dos ciclos de telemetria armazenados.
func (r *Repository) GetTelemetryAnalytics() (TelemetryAnalytics, error) {
	var analytics TelemetryAnalytics

	records, err := r.GetRecentTelemetryRecords(12)
	if err != nil {
		return analytics, err
	}

	analytics.TotalRecords = len(records)
	if analytics.TotalRecords == 0 {
		return analytics, nil
	}

	// Calcula medias moveis de IDT
	var sum3, sum6, sum12 float64
	n := len(records)

	for i, rec := range records {
		if i >= n-3 {
			sum3 += rec.IDT
		}
		if i >= n-6 {
			sum6 += rec.IDT
		}
		sum12 += rec.IDT
	}

	if n >= 3 {
		analytics.IDT3 = p2t.RoundPercentage(sum3 / 3.0)
	} else {
		analytics.IDT3 = p2t.RoundPercentage(sum3 / float64(n))
	}

	count6 := 6
	if n < 6 {
		count6 = n
	}
	analytics.IDT6 = p2t.RoundPercentage(sum6 / float64(count6))

	analytics.IDT12 = p2t.RoundPercentage(sum12 / float64(n))

	// Calcula tendencia do VRH (comparando o ultimo com a media anterior)
	if n >= 2 {
		lastVRH := records[n-1].VRH
		prevVRH := records[n-2].VRH
		if prevVRH > 0 {
			analytics.VRHTrendPercent = p2t.RoundPercentage(((lastVRH - prevVRH) / prevVRH) * 100.0)
		}
	}

	// Soma de custos invisiveis
	querySum := `SELECT COALESCE(SUM(invisible_costs), 0) FROM telemetry_cycles;`
	err = r.db.QueryRow(querySum).Scan(&analytics.TotalAnnualInvisibleCosts)
	if err != nil {
		return analytics, fmt.Errorf("falha ao calcular total de custos invisiveis: %w", err)
	}
	analytics.TotalAnnualInvisibleCosts = p2t.RoundCurrency(analytics.TotalAnnualInvisibleCosts)

	return analytics, nil
}

// DeleteTelemetryRecord remove um registro de telemetria pelo ID.
func (r *Repository) DeleteTelemetryRecord(id int64) error {
	query := `DELETE FROM telemetry_cycles WHERE id = ?;`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao excluir registro de telemetria #%d: %w", id, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("registro de telemetria #%d nao encontrado", id)
	}
	return nil
}

// UpdateTelemetryRecord atualiza um registro de telemetria existente pelo ID.
func (r *Repository) UpdateTelemetryRecord(id int64, input p2t.TelemetryInput, res p2t.TelemetryResult, referenceMonth string) error {
	query := `
	UPDATE telemetry_cycles
	SET gross_salary = ?, fixed_deductions = ?, error_deductions = ?, invisible_costs = ?,
	    contract_hours = ?, commute_hours = ?, total_hours = ?, real_liquidity = ?,
	    vrh = ?, idt = ?, reference_month = ?
	WHERE id = ?;
	`
	result, err := r.db.Exec(query,
		input.GrossSalary, input.FixedDeductions, input.ErrorDeductions, input.InvisibleCosts,
		input.ContractHours, input.CommuteHours, res.TotalHours, res.RealLiquidity,
		res.VRH, res.IDT, referenceMonth, id,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar registro de telemetria #%d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("registro de telemetria #%d nao encontrado", id)
	}
	return nil
}

// DeleteBufferRecord remove um registro de ciclo de buffer pelo ID.
func (r *Repository) DeleteBufferRecord(id int64) error {
	query := `DELETE FROM buffer_cycles WHERE id = ?;`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("falha ao excluir ciclo de buffer #%d: %w", id, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("ciclo de buffer #%d nao encontrado", id)
	}
	return nil
}


