package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
)

// GoalRecord representa uma meta persistida.
type GoalRecord struct {
	p2t.Goal
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RecurringRecord representa um compromisso recorrente persistido.
type RecurringRecord struct {
	p2t.RecurringCommitment
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GoalSnapshotProgressRecord representa o progresso de uma meta no fechamento.
type GoalSnapshotProgressRecord struct {
	GoalID              int64
	CurrentBalance      float64
	PlannedContribution float64
}

// FinancialSnapshotRecord representa um fechamento mensal e o progresso das suas metas.
type FinancialSnapshotRecord struct {
	p2t.FinancialSnapshot
	Goals []GoalSnapshotProgressRecord
}

// CreateGoal cria uma meta ativa.
func (r *Repository) CreateGoal(goal p2t.Goal) (int64, error) {
	goal.Archived = false
	if err := goal.Validate(); err != nil {
		return 0, fmt.Errorf("meta invalida: %w", err)
	}
	result, err := r.db.Exec(`
		INSERT INTO goals (name, target_amount, deadline, current_balance, monthly_contribution, archived)
		VALUES (?, ?, ?, ?, ?, 0);`,
		goal.Name, goal.TargetAmount, goal.Deadline, goal.CurrentBalance, goal.MonthlyContribution,
	)
	if err != nil {
		return 0, fmt.Errorf("falha ao criar meta: %w", err)
	}
	return result.LastInsertId()
}

// GetGoal busca uma meta pelo ID.
func (r *Repository) GetGoal(id int64) (*GoalRecord, error) {
	var record GoalRecord
	err := r.db.QueryRow(`
		SELECT id, name, target_amount, deadline, current_balance, monthly_contribution,
		       archived, created_at, updated_at
		FROM goals WHERE id = ?;`, id).Scan(
		&record.ID, &record.Name, &record.TargetAmount, &record.Deadline,
		&record.CurrentBalance, &record.MonthlyContribution, &record.Archived,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar meta #%d: %w", id, err)
	}
	return &record, nil
}

// ListGoals lista metas ativas ou todas quando includeArchived for verdadeiro.
func (r *Repository) ListGoals(includeArchived bool) ([]GoalRecord, error) {
	query := `
		SELECT id, name, target_amount, deadline, current_balance, monthly_contribution,
		       archived, created_at, updated_at
		FROM goals`
	if !includeArchived {
		query += ` WHERE archived = 0`
	}
	query += ` ORDER BY deadline ASC, id ASC;`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar metas: %w", err)
	}
	defer rows.Close()

	var records []GoalRecord
	for rows.Next() {
		var record GoalRecord
		if err := rows.Scan(
			&record.ID, &record.Name, &record.TargetAmount, &record.Deadline,
			&record.CurrentBalance, &record.MonthlyContribution, &record.Archived,
			&record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("falha ao ler meta: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// UpdateGoal atualiza todos os dados editáveis de uma meta.
func (r *Repository) UpdateGoal(goal p2t.Goal) error {
	if goal.ID <= 0 {
		return fmt.Errorf("id da meta deve ser maior que zero")
	}
	if err := goal.Validate(); err != nil {
		return fmt.Errorf("meta invalida: %w", err)
	}
	result, err := r.db.Exec(`
		UPDATE goals
		SET name = ?, target_amount = ?, deadline = ?, current_balance = ?,
		    monthly_contribution = ?, archived = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?;`,
		goal.Name, goal.TargetAmount, goal.Deadline, goal.CurrentBalance,
		goal.MonthlyContribution, goal.Archived, goal.ID,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar meta #%d: %w", goal.ID, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("meta #%d nao encontrada", goal.ID)
	}
	return nil
}

// ArchiveGoal arquiva uma meta sem apagar seu histórico.
func (r *Repository) ArchiveGoal(id int64) error {
	result, err := r.db.Exec(`UPDATE goals SET archived = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("falha ao arquivar meta #%d: %w", id, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("meta #%d nao encontrada", id)
	}
	return nil
}

// CreateRecurring cria um compromisso recorrente ativo.
func (r *Repository) CreateRecurring(commitment p2t.RecurringCommitment) (int64, error) {
	commitment.Status = p2t.RecurringActive
	if err := commitment.Validate(); err != nil {
		return 0, fmt.Errorf("compromisso invalido: %w", err)
	}
	result, err := r.db.Exec(`
		INSERT INTO recurring_commitments
		(name, amount, period, purpose, essentiality, billing_day, status, last_reviewed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);`,
		commitment.Name, commitment.Amount, commitment.Period, commitment.Purpose,
		commitment.Essentiality, commitment.BillingDay, commitment.Status, commitment.LastReviewed,
	)
	if err != nil {
		return 0, fmt.Errorf("falha ao criar compromisso: %w", err)
	}
	return result.LastInsertId()
}

// GetRecurring busca um compromisso pelo ID.
func (r *Repository) GetRecurring(id int64) (*RecurringRecord, error) {
	var record RecurringRecord
	err := r.db.QueryRow(`
		SELECT id, name, amount, period, purpose, essentiality, billing_day, status,
		       last_reviewed, created_at, updated_at
		FROM recurring_commitments WHERE id = ?;`, id).Scan(
		&record.ID, &record.Name, &record.Amount, &record.Period, &record.Purpose,
		&record.Essentiality, &record.BillingDay, &record.Status, &record.LastReviewed,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar compromisso #%d: %w", id, err)
	}
	return &record, nil
}

// ListRecurring lista compromissos ativos ou todos quando includeArchived for verdadeiro.
func (r *Repository) ListRecurring(includeArchived bool) ([]RecurringRecord, error) {
	query := `
		SELECT id, name, amount, period, purpose, essentiality, billing_day, status,
		       last_reviewed, created_at, updated_at
		FROM recurring_commitments`
	if !includeArchived {
		query += ` WHERE status = 'active'`
	}
	query += ` ORDER BY name ASC, id ASC;`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar compromissos: %w", err)
	}
	defer rows.Close()

	var records []RecurringRecord
	for rows.Next() {
		var record RecurringRecord
		if err := rows.Scan(
			&record.ID, &record.Name, &record.Amount, &record.Period, &record.Purpose,
			&record.Essentiality, &record.BillingDay, &record.Status, &record.LastReviewed,
			&record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("falha ao ler compromisso: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// GetActiveRecurring retorna compromissos que entram no cálculo do fechamento.
func (r *Repository) GetActiveRecurring() ([]RecurringRecord, error) {
	return r.ListRecurring(false)
}

// UpdateRecurring atualiza todos os dados editáveis de um compromisso.
func (r *Repository) UpdateRecurring(commitment p2t.RecurringCommitment) error {
	if commitment.ID <= 0 {
		return fmt.Errorf("id do compromisso deve ser maior que zero")
	}
	if err := commitment.Validate(); err != nil {
		return fmt.Errorf("compromisso invalido: %w", err)
	}
	result, err := r.db.Exec(`
		UPDATE recurring_commitments
		SET name = ?, amount = ?, period = ?, purpose = ?, essentiality = ?,
		    billing_day = ?, status = ?, last_reviewed = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?;`,
		commitment.Name, commitment.Amount, commitment.Period, commitment.Purpose,
		commitment.Essentiality, commitment.BillingDay, commitment.Status,
		commitment.LastReviewed, commitment.ID,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar compromisso #%d: %w", commitment.ID, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("compromisso #%d nao encontrado", commitment.ID)
	}
	return nil
}

// ArchiveRecurring arquiva um compromisso sem apagar seu histórico.
func (r *Repository) ArchiveRecurring(id int64) error {
	result, err := r.db.Exec(`UPDATE recurring_commitments SET status = 'archived', updated_at = CURRENT_TIMESTAMP WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("falha ao arquivar compromisso #%d: %w", id, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("compromisso #%d nao encontrado", id)
	}
	return nil
}

// SaveFinancialSnapshot cria ou atualiza o fechamento mensal e seus progressos de metas.
func (r *Repository) SaveFinancialSnapshot(snapshot p2t.FinancialSnapshot, progress []GoalSnapshotProgressRecord) (int64, error) {
	if err := snapshot.Validate(); err != nil {
		return 0, fmt.Errorf("fechamento invalido: %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("falha ao iniciar fechamento: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO financial_snapshots (
			reference_month, received_income, reserve_balance, planned_goal_contributions,
			exceptional_costs, recurring_commitment_total, goal_capacity, free_margin
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(reference_month) DO UPDATE SET
			received_income = excluded.received_income,
			reserve_balance = excluded.reserve_balance,
			planned_goal_contributions = excluded.planned_goal_contributions,
			exceptional_costs = excluded.exceptional_costs,
			recurring_commitment_total = excluded.recurring_commitment_total,
			goal_capacity = excluded.goal_capacity,
			free_margin = excluded.free_margin,
			updated_at = CURRENT_TIMESTAMP;`,
		snapshot.ReferenceMonth, snapshot.ReceivedIncome, snapshot.ReserveBalance,
		snapshot.PlannedGoalContributions, snapshot.ExceptionalCosts,
		snapshot.RecurringCommitmentTotal, snapshot.GoalCapacity, snapshot.FreeMargin,
	)
	if err != nil {
		return 0, fmt.Errorf("falha ao salvar fechamento: %w", err)
	}

	var snapshotID int64
	if err := tx.QueryRow(`SELECT id FROM financial_snapshots WHERE reference_month = ?;`, snapshot.ReferenceMonth).Scan(&snapshotID); err != nil {
		return 0, fmt.Errorf("falha ao obter id do fechamento: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM goal_snapshot_progress WHERE snapshot_id = ?;`, snapshotID); err != nil {
		return 0, fmt.Errorf("falha ao substituir progresso das metas: %w", err)
	}
	for _, item := range progress {
		if item.GoalID <= 0 || item.CurrentBalance < 0 || item.PlannedContribution < 0 {
			return 0, fmt.Errorf("progresso de meta invalido para a meta #%d", item.GoalID)
		}
		if _, err := tx.Exec(`
			INSERT INTO goal_snapshot_progress (snapshot_id, goal_id, current_balance, planned_contribution)
			VALUES (?, ?, ?, ?);`, snapshotID, item.GoalID, item.CurrentBalance, item.PlannedContribution); err != nil {
			return 0, fmt.Errorf("falha ao salvar progresso da meta #%d: %w", item.GoalID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("falha ao confirmar fechamento: %w", err)
	}
	return snapshotID, nil
}

// GetFinancialSnapshot busca um fechamento mensal e o progresso das suas metas.
func (r *Repository) GetFinancialSnapshot(referenceMonth string) (*FinancialSnapshotRecord, error) {
	if err := p2t.ValidateReferenceMonth(referenceMonth); err != nil {
		return nil, err
	}
	var record FinancialSnapshotRecord
	err := r.db.QueryRow(`
		SELECT id, reference_month, received_income, reserve_balance, planned_goal_contributions,
		       exceptional_costs, recurring_commitment_total, goal_capacity, free_margin, created_at
		FROM financial_snapshots WHERE reference_month = ?;`, referenceMonth).Scan(
		&record.ID, &record.ReferenceMonth, &record.ReceivedIncome, &record.ReserveBalance,
		&record.PlannedGoalContributions, &record.ExceptionalCosts,
		&record.RecurringCommitmentTotal, &record.GoalCapacity, &record.FreeMargin,
		&record.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar fechamento %s: %w", referenceMonth, err)
	}

	rows, err := r.db.Query(`
		SELECT goal_id, current_balance, planned_contribution
		FROM goal_snapshot_progress
		WHERE snapshot_id = ? ORDER BY goal_id ASC;`, record.ID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar progresso do fechamento: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item GoalSnapshotProgressRecord
		if err := rows.Scan(&item.GoalID, &item.CurrentBalance, &item.PlannedContribution); err != nil {
			return nil, fmt.Errorf("falha ao ler progresso do fechamento: %w", err)
		}
		record.Goals = append(record.Goals, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &record, nil
}
